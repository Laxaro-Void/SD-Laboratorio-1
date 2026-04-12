package main

import (
	"os"
	"log"
	"fmt"
	"sync"

	pb "EquiposNinja/proto/message"
	pbDataAkatsuki "EquiposNinja/proto/DataAkatsuki"
	"google.golang.org/protobuf/proto"

	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func dialRabbitMQ() (*amqp091.Connection, error) {
	log.Printf("Connecting to RabbitMQ at %s", "amqp://guest:guest@" + os.Getenv("RABBITMQ-IP") + "/")
	conn, err := amqp091.Dial("amqp://guest:guest@" + os.Getenv("RABBITMQ-IP") + "/")
	if err != nil {
		return nil, err
	}
	return conn, nil
}

type Equipo struct {
	Nombre   string
	Ataque   int
	Vida     int
	BolsaRyo float64
}

type Akatsuki struct {
	Id int
	Name string
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

func (al *AkatsukiList) ReplaceALLAkatsuki(akatsuki []Akatsuki)  {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.Akatsukis = make(map[int]Akatsuki)
	for _, a := range akatsuki {
		al.Akatsukis[a.Id] = a
	}
}

func listeningInfoAnbu() {
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
		// Aquí se puede agregar lógica para procesar el mensaje y actualizar el estado del equipo ninja
	}
}

func mostrarEstadisticas(equipo *Equipo) {
	log.Printf("Nombre: %s", 	   equipo.Nombre  )
	log.Printf("Ataque: %d", 	   equipo.Ataque  )
	log.Printf("Vida: %d", 		   equipo.Vida    )
	log.Printf("Bolsa de Ryo: %.2f", equipo.BolsaRyo)
}

func solicitarLista(clientHokage pb.HokageClient, equipo *Equipo) {
	
}

func solicitarRecompensa(clientHokage pb.HokageClient, equipo *Equipo) {

}

func atacarAkatsuki(clientAkatsuki pb.AkatsukiClient, equipo *Equipo) {
	fmt.Println("Iniciando simulación de combate...")
	var nombre string
	fmt.Print("Ingrese Nombre del enemigo que desea atacar: ")
	fmt.Scanln((&nombre))
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

	go listeningInfoAnbu()

	// ----------------------------- Conexiones ----------------------------- //
	// 1. Conexión gRPC con Hokage (MV1)
	connHokage, err := grpc.NewClient("" + os.Getenv("Hokage-IP"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("No se pudo conectar con Hokage: %v", err)
		return
	}
	defer connHokage.Close()
	clientHokage := pb.NewHokageClient(connHokage)

	// 2. Conexión gRPC con Akatsuki (MV3)
	connAkatsuki, err := grpc.NewClient("" + os.Getenv("Akatsuki-IP"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("No se pudo conectar con Akatsuki: %v", err)
	}
	defer connAkatsuki.Close()
	clientAkatsuki := pb.NewAkatsukiClient(connAkatsuki)

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
			solicitarLista(clientHokage, statsEquipo)
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
