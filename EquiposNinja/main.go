package main

import (
	"os"
	"log"
	"fmt"
	"sync"
	"context"
	"time"

	pbDataHokage "EquiposNinja/proto/DataHokage"
	pbDataAkatsuki "EquiposNinja/proto/DataAkatsuki"
	pbCombateAkatsuki "EquiposNinja/proto/CombateAkatsuki"
	"google.golang.org/protobuf/proto"

	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

func dialRabbitMQ() (*amqp091.Connection, error) {
	var maxAttempts = 10
	var attempt int
	for attempt = 1; attempt <= maxAttempts; attempt++ {
		log.Printf("Attempting to connect to RabbitMQ (Attempt %d/%d)", attempt, maxAttempts)
		conn, err := amqp091.Dial("amqp://user:pass@" + os.Getenv("RABBITMQ-IP") + "/")
		if err == nil {
			log.Printf("Successfully connected to RabbitMQ on attempt %d", attempt)
			return conn, nil
		}
		log.Printf("Failed to connect to RabbitMQ on attempt %d: %v", attempt, err)
		time.Sleep(5 * time.Second) // Wait before retrying
	}
	return nil, fmt.Errorf("could not connect to RabbitMQ after %d attempts", maxAttempts)
}

type Equipo struct {
	Nombre   string
	Ataque   int
	Vida     int
	BolsaRyo float64
}

type Akatsuki struct {
	Id int
	Nombre string
	Ataque int
	Vida int
	Estado string
	isComunicated bool
}

type AkatsukiList struct {
	Akatsukis map[int]Akatsuki
	mu sync.Mutex
}

func (al *AkatsukiList) AddAkatsuki(akatsuki Akatsuki) {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.Akatsukis[akatsuki.Id] = akatsuki
}

func (al *AkatsukiList) GetAkatsuki(id int) (Akatsuki, bool) {
	al.mu.Lock()
	defer al.mu.Unlock()
	akatsuki, exists := al.Akatsukis[id]
	return akatsuki, exists
}

func (al *AkatsukiList) ReplaceAkatsuki(akatsuki Akatsuki)  {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.Akatsukis[akatsuki.Id] = akatsuki
}

func (al *AkatsukiList) ReplaceEstado(id int, estado string) {
	al.mu.Lock()
	defer al.mu.Unlock()
	if akatsuki, exists := al.Akatsukis[id]; exists {
		akatsuki.Estado = estado
		al.Akatsukis[id] = akatsuki
	}
}

func (al *AkatsukiList) ReplaceALLAkatsuki(akatsuki []Akatsuki)  {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.Akatsukis = make(map[int]Akatsuki)
	for _, a := range akatsuki {
		al.Akatsukis[a.Id] = a
	}
}

func listeningInfoAnbu(akatsukiList *AkatsukiList) {
	// Conexión RabbitMQ para recibir información de Anbu
	connRabbit, err := dialRabbitMQ()
	if err != nil {
		log.Fatalf("Error RabbitMQ: %v", err)
	}
	defer connRabbit.Close()
	ch, err := connRabbit.Channel()
	if err != nil {
		log.Fatalf("Error al abrir canal RabbitMQ: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"",    // nombre vacío → cola exclusiva
		false,
		true,
		true,
		false,
		nil,
	)
	err = ch.QueueBind(
		q.Name,
		"",
		"EquiposNinjaBrodcast",
		false,
		nil,
	)

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	for d := range msgs {
		var data pbDataAkatsuki.DataAkatsukiRequest
		err := proto.Unmarshal(d.Body, &data)
		if err != nil {
			log.Printf("Error al deserializar mensaje: %v", err)
			continue
		}
		log.Printf("Mensaje recibido de Anbu: %s", data.Nombre)
		akatsuki := Akatsuki{
			Id:     int(data.Id),
			Estado: data.Estado,
		}
		if _, exists := akatsukiList.GetAkatsuki(akatsuki.Id); exists {
			akatsukiList.ReplaceEstado(akatsuki.Id, akatsuki.Estado)
		}
	}
}

func mostrarEstadisticas(equipo *Equipo) {
	log.Printf("Nombre: %s", 	   equipo.Nombre  )
	log.Printf("Ataque: %d", 	   equipo.Ataque  )
	log.Printf("Vida: %d", 		   equipo.Vida    )
	log.Printf("Bolsa de Ryo: %.2f", equipo.BolsaRyo)
}

func solicitarLista(client pbDataHokage.HokageClient, akatsukiList *AkatsukiList) {
	response, err:= client.ObtenerListaAkatsukis(context.Background(),&pbDataHokage.Empty{})
	if err != nil {
		log.Printf("Error al solicitar lista de Akatsukis: %v", err)
		return
	}
	log.Printf("Lista de Akatsukis:")
	for _, akatsuki := range response.Akatsukis {
		fmt.Printf("ID: %d / Nombre: %s / Ataque: %d / Vida: %d / Estado: %s / Recompensa: %d", akatsuki.Id, akatsuki.Nombre, akatsuki.Ataque, akatsuki.Vida, akatsuki.Estado, akatsuki.Recompensa)
		akatsukiList.ReplaceAkatsuki(Akatsuki{Id: int(akatsuki.Id), Nombre: akatsuki.Nombre, Ataque: int(akatsuki.Ataque), Vida: int(akatsuki.Vida), Estado: akatsuki.Estado, isComunicated: false})
	}
}

func solicitarRecompensa(clientHokage pbDataHokage.HokageClient, equipo *Equipo) {

}

func atacarAkatsuki(client pbCombateAkatsuki.CombateAkatsukiClient, equipo *Equipo) {
	fmt.Println("Iniciando simulación de combate...")
	var nombre string
	fmt.Print("Ingrese Nombre del enemigo que desea atacar: ")
	fmt.Scanln((&nombre))

	response, err := client.IniciarCombate(
		context.Background(), 
		&pbCombateAkatsuki.IniciarCombateRequest{
			Equipo: &pbCombateAkatsuki.DatosEquipo{NombreEquipo: equipo.Nombre, Ataque: int32(equipo.Ataque), Vida: int32(equipo.Vida)}, 
			IdObjetivo: 0, 
			NombreObjetivo: ""})
	
	if err != nil {
		log.Printf("Error al obtener resultados de combate: %v", err)
		return
	}

	log.Printf("Resultado %v", response.Success)
}

func crearEquipoNinja() {
	statsEquipo := &Equipo{
		Nombre:   "",
		Ataque:   0,
		Vida:     0,
		BolsaRyo: 0,
	}
	
	log.Print("Ingrese el nombre de su equipo ninja: ")
	fmt.Scanf("%s", &statsEquipo.Nombre)
	log.Print("Ingrese el ataque de su equipo ninja: ")
	fmt.Scanf("%d", &statsEquipo.Ataque)
	log.Print("Ingrese la vida de su equipo ninja: ")
	fmt.Scanf("%d", &statsEquipo.Vida)
	log.Print("Ingrese la bolsa de ryo de su equipo ninja: ")
	fmt.Scanf("%f", &statsEquipo.BolsaRyo)

	log.Printf("Equipo ninja creado: %s", statsEquipo.Nombre)
	log.Printf("Ataque: %d", statsEquipo.Ataque)
	log.Printf("Vida: %d", statsEquipo.Vida)
	log.Printf("Bolsa de Ryo: %.2f", statsEquipo.BolsaRyo)

	akatsukiList := &AkatsukiList{
		Akatsukis: make(map[int]Akatsuki),
	}

	go listeningInfoAnbu(akatsukiList)

	// ----------------------------- Conexiones ----------------------------- //
	// 1. Conexión gRPC con Hokage (MV1)
	connHokage, err := grpc.Dial(os.Getenv("Hokage-IP"), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("No se pudo conectar con Hokage: %v", err)
		return
	}
	defer connHokage.Close()
	log.Printf("Conectado a Hokage en %s", connHokage.Target())
	clientHokage := pbDataHokage.NewHokageClient(connHokage)

	// 2. Conexión gRPC con Akatsuki (MV3)
	connAkatsuki, err := grpc.Dial(os.Getenv("Akatsuki-IP"), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("No se pudo conectar con Akatsuki: %v", err)
	}
	defer connAkatsuki.Close()
	log.Printf("Conectado a Akatsuki en %s", connAkatsuki.Target())
	clientAkatsuki := pbCombateAkatsuki.NewCombateAkatsukiClient(connAkatsuki)

	// ------------------------------ Interfaz de Usuario ----------------------------- //
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
			// Lógica para pedir lista
			solicitarLista(clientHokage, akatsukiList)
		case 2:
			// Lógica para atacar Akatsuki
			atacarAkatsuki(clientAkatsuki, statsEquipo)
		case 3:
			// Lógica para solicitar recompensa
			solicitarRecompensa(clientHokage, statsEquipo)
		case 4:
			// Lógica para ver estadísticas
			mostrarEstadisticas(statsEquipo)
		default:
			log.Println("Opción no válida")
		}
	}
}


func main() {
	name := os.Getenv("NAME")
	port := os.Getenv("PORT")
	log.Printf("%s is running on port %s", name, port)

	crearEquipoNinja()
}
