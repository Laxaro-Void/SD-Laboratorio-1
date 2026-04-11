package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	pb "EquiposNinja/proto/message"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Equipo struct {
	Nombre   string
	Ataque   int
	Vida     int
	BolsaRyo float64
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var nombre string
	var atk, vida int
	var ryo float64
	fmt.Println("Nuevo Equipo!")
	fmt.Print("Ingrese Nombre: ")
	fmt.Scanln(&nombre)
	fmt.Print("\nIngrese Ataque: ")
	fmt.Scanln(&atk)
	fmt.Print("\nIngrese Vida: ")
	fmt.Scanln(&vida)
	fmt.Print("\nSaldo Ryo: ")
	fmt.Scanln(&ryo)

	miEquipo := &Equipo{
		Nombre:   nombre,
		Ataque:   atk,
		Vida:     vida,
		BolsaRyo: ryo,
	}

	var mu sync.Mutex
	esperaANBU := make(chan bool)

	// 1. Conexión gRPC con Hokage (MV1)
	connHokage, err := grpc.NewClient("hokage-vm:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("No se pudo conectar con Hokage: %v", err)
	}
	defer connHokage.Close()
	clientHokage := pb.NewHokageClient(connHokage)

	// 2. Conexión RabbitMQ con ANBU (MV2)
	connRabbit, err := amqp.Dial("amqp://guest:guest@anbu-vm:5672/")
	if err != nil {
		log.Fatalf("Error RabbitMQ: %v", err)
	}
	defer connRabbit.Close()

	// 3. Conexión gRPC con Akatsuki (MV3)
	connAkatsuki, err := grpc.NewClient("akatsuki-vm:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("No se pudo conectar con Hokage: %v", err)
	}
	defer connHokage.Close()
	clientAkatsuki := pb.NewAkatsukiClient(connAkatsuki)

	ch, _ := connRabbit.Channel()
	msgs, _ := ch.Consume("alertas_anbu", "", true, false, false, false, nil)

	// Goroutine para escuchar a ANBU asíncronamente
	go func() {
		for d := range msgs {
			mu.Lock()
			// 1. Imprimimos el reporte en pantalla inmediatamente
			fmt.Printf("\n[ANBU Alerta] %s\n", d.Body)
			mu.Unlock()
			esperaANBU <- true
		}
	}()

	// 3. Bucle de interacción del usuario
	for {
		fmt.Println("\n======= ACCIONES ========")
		fmt.Println("1. Pedir Lista")
		fmt.Println("2. Atacar Akatsuki")
		fmt.Println("3. Solicitar Recompensa")
		fmt.Println("4. Ver estadísticas")
		fmt.Println("==========================")

		var opcion int
		fmt.Scanln(&opcion)

		switch opcion {
		case 1:
			res, _ := clientHokage.ObtenerListaAkatsukis(context.Background(), &pb.Empty{})
			fmt.Println("Lista de Akatsukis")
			for _, a := range res.Enemigos {
				fmt.Printf("%s / Ataque = %s / Vida = %s / Estado = %s / Recompensa = %s\n", a.Nombre, a.Ataque, a.Vida, a.Estado, a.Recompensa)
			}

		case 2:
			// Aquí se llamaría a la entidad Akatsuki para iniciar combate [cite: 112]
			fmt.Println("Iniciando simulación de combate...")
			var nombre string
			fmt.Print("Ingrese Nombre del enemigo que desea atacar: ")
			fmt.Scanln((&nombre))
			entrada := pb.DatosEquipo{
				NombreEquipo: nombre,
				Ataque:       int32(miEquipo.Ataque),
				Vida:         int32(miEquipo.Vida),
			}

			mu.Lock()
			res, err := clientAkatsuki.IniciarCombate(context.Background(), &entrada)
			if err != nil {
				fmt.Printf("Error de conexión con Akatsuki: %v\n", err)
				continue
			}
			fmt.Printf("Estado de %s: %s\n", nombre, res.EstadoFinal)

			if res.CombateIniciado == true {
				fmt.Println("Comienza el Combate!!!")
				mu.Unlock()
				mu.Lock()
			}

			mu.Unlock()

		case 3:

			return

		case 4:
			fmt.Printf("Nombre: %s", miEquipo.Nombre)
			fmt.Printf("Ataque: %s", miEquipo.Ataque)
			fmt.Printf("Vida: %s", miEquipo.Vida)
			fmt.Printf("Bolsa Ryo: %sM", miEquipo.BolsaRyo)
		}
	}
}
