package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	pbCombateAkatsuki "EquiposNinja/proto/CombateAkatsuki"
	pbDataAkatsuki "EquiposNinja/proto/DataAkatsuki"
	pbDataHokage "EquiposNinja/proto/DataHokage"

	"google.golang.org/protobuf/proto"

	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

/*
	dialRabbitMQ() (*amqp091.Connection, error)
	- Intenta establecer una conexión con RabbitMQ, reintentando varias veces en caso de fallo.
	- Retorna la conexión establecida o un error si no se pudo conectar después de varios intentos.
*/
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

/*
	AkatsukiList struct
	- Almacena de manera protegida informacion de los Akatsukis detectados.
*/
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

/*
	listeningInfoAnbu(akatsukiList *AkatsukiList)

	- Establece una conexión con RabbitMQ para recibir información de Anbu sobre los Akatsukis detectados.
*/
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
		akatsuki := Akatsuki{
			Id:     int(data.Id),
			Estado: data.Estado,
		}
		if _, exists := akatsukiList.GetAkatsuki(akatsuki.Id); exists {
			// Actualizar estado del Akatsuki en la lista local
			akatsukiList.ReplaceEstado(akatsuki.Id, akatsuki.Estado)
		} else {
			// Agregar nuevo Akatsuki a la lista local
			akatsukiList.AddAkatsuki(akatsuki)
		}
	}
}

/*
	mostrarEstadisticas(equipo *Equipo)

	- Muestra las estadísticas del equipo ninja, incluyendo nombre, ataque, vida y bolsa de Ryo.
*/
func mostrarEstadisticas(equipo *Equipo) {
	fmt.Printf("Nombre: %s", 	   equipo.Nombre  )
	fmt.Printf("Ataque: %d", 	   equipo.Ataque  )
	fmt.Printf("Vida: %d", 		   equipo.Vida    )
	fmt.Printf("Bolsa de Ryo: %.2f", equipo.BolsaRyo)
}

/*
	solicitarLista(client pbDataHokage.HokageClient, akatsukiList *AkatsukiList)

	- Solicita la lista de Akatsukis a Hokage a través de gRPC y actualiza la lista local con la información recibida.
	- Muestra la lista de Akatsukis obtenida de Hokage.
*/
func solicitarLista(client pbDataHokage.HokageClient, akatsukiList *AkatsukiList) {
	response, err:= client.ObtenerListaAkatsukis(context.Background(),&pbDataHokage.Empty{})
	if err != nil {
		log.Printf("Error al solicitar lista de Akatsukis: %v", err)
		return
	}
	log.Printf("Lista de Akatsukis:")
	for _, akatsuki := range response.Akatsukis {
		fmt.Printf("ID: %d / Nombre: %s / Ataque: %d / Vida: %d / Estado: %s / Recompensa: %d\n", akatsuki.Id, akatsuki.Nombre, akatsuki.Ataque, akatsuki.Vida, akatsuki.Estado, akatsuki.Recompensa)
		akatsukiList.ReplaceAkatsuki(Akatsuki{Id: int(akatsuki.Id), Nombre: akatsuki.Nombre, Ataque: int(akatsuki.Ataque), Vida: int(akatsuki.Vida), Estado: akatsuki.Estado, isComunicated: false})
	}
}

/*
	solicitarRecompensa(clientHokage pbDataHokage.HokageClient, clienteAkatsuki pbCombateAkatsuki.CombateAkatsukiClient, equipo *Equipo)

	- Solicita la recompensa por capturar un Akatsuki a Hokage a través de gRPC.
	- Si la recompensa es otorgada, actualiza la bolsa de Ryo del equipo ninja.
	- Luego, envía una solicitud a Akatsuki para liberar al Akatsuki capturado.
*/
func solicitarRecompensa(clientHokage pbDataHokage.HokageClient, clienteAkatsuki pbCombateAkatsuki.CombateAkatsukiClient, equipo *Equipo) {
	log.Print("Ingrese el nombre del Akatsuki por el que desea solicitar la recompensa: ")
	var nombreAkatsuki string
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		nombreAkatsuki = scanner.Text()
	}

	response, err := clientHokage.ReclamarRecompensa(context.Background(), &pbDataHokage.SolicitudRecompensa{NombreEquipo: equipo.Nombre, NombreAkatsuki: nombreAkatsuki})
	if err != nil {
		log.Printf("%s", strings.Split(err.Error(), "= ")[2])
		return
	}

	if response.RyoPagados > 0 {
		equipo.BolsaRyo += float64(response.RyoPagados)
		log.Printf("¡Recompensa recibida! Ryo pagados: %d. Bolsa de Ryo actualizada: %.2f", response.RyoPagados, equipo.BolsaRyo)
	} else {
		log.Printf("No se pudo obtener la recompensa para el Akatsuki %s. Mensaje: %s", nombreAkatsuki, response.Mensaje)
	}

	// Permite "liberar" al Akatsuki capturado para que vuelva a estar disponible en la lista de Akatsukis y pueda ser atacado nuevamente por otros equipos, si localizado nuevamente
	_, err = clienteAkatsuki.LiberarAkatsuki(context.Background(), &pbCombateAkatsuki.LiberarAkatsukiRequest{NombreAkatsuki: nombreAkatsuki})
	if err != nil {
		log.Printf("Error al liberar Akatsuki: %v", err)
		return
	}
}

/*
	atarcarAkatsuki(client pbCombateAkatsuki.CombateAkatsukiClient, equipo *Equipo)

	- Permite al equipo ninja atacar a un Akatsuki especificando su nombre.
	- Envía una solicitud a Akatsuki para iniciar el combate y muestra el resultado (victoria o derrota) al usuario.
*/
func atacarAkatsuki(client pbCombateAkatsuki.CombateAkatsukiClient, equipo *Equipo) {
	fmt.Println("Iniciando simulación de combate...")
	var nombre string
	fmt.Print("Ingrese Nombre del enemigo que desea atacar: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		nombre = scanner.Text()
	}

	response, err := client.IniciarCombate(
		context.Background(), 
		&pbCombateAkatsuki.IniciarCombateRequest{
			Equipo: &pbCombateAkatsuki.DatosEquipo{NombreEquipo: equipo.Nombre, Ataque: int32(equipo.Ataque), Vida: int32(equipo.Vida)}, 
			IdObjetivo: 0, 
			NombreObjetivo: nombre})
	
	if err != nil {
		log.Printf("%s", strings.Split(err.Error(), "= ")[2])
		return
	}

	if response.Success {
		log.Printf("¡Victoria! El Akatsuki ha sido derrotado, Listo para recibir recompensa.")
	} else {
		log.Printf("Derrota. El Akatsuki es demasiado fuerte.")
	}
}

/*
	crearEquipoNinja()

	- Permite al usuario crear su equipo ninja ingresando el nombre, ataque, vida y bolsa de Ryo.
	- Luego, establece las conexiones necesarias con Hokage y Akatsuki, y muestra un menú de acciones para interactuar con el sistema.
*/
func crearEquipoNinja() {
	statsEquipo := &Equipo{
		Nombre:   "",
		Ataque:   0,
		Vida:     0,
		BolsaRyo: 0,
	}
	
	log.Print("Ingrese el nombre de su equipo ninja: ")
	for {
		fmt.Scanf("%s", &statsEquipo.Nombre)
		if statsEquipo.Nombre != "" {
			break
		}
		log.Print("El nombre no puede estar vacío. Ingrese el nombre de su equipo ninja: ")
	}
	log.Print("Ingrese el ataque de su equipo ninja: ")
	for {
		fmt.Scanf("%d", &statsEquipo.Ataque)
		if statsEquipo.Ataque > 0 {
			break
		}
		log.Print("El ataque debe ser un número positivo. Ingrese el ataque de su equipo ninja: ")
	}
	log.Print("Ingrese la vida de su equipo ninja: ")
	for {
		fmt.Scanf("%d", &statsEquipo.Vida)
		if statsEquipo.Vida > 0 {
			break
		}
		log.Print("La vida debe ser un número positivo. Ingrese la vida de su equipo ninja: ")
	}
	log.Print("Ingrese la bolsa de ryo de su equipo ninja: ")
	for {
		fmt.Scanf("%f", &statsEquipo.BolsaRyo)
		if statsEquipo.BolsaRyo > 0 {
			break
		}
		log.Print("La bolsa de ryo debe ser un número positivo. Ingrese la bolsa de ryo de su equipo ninja: ")
	}

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
			solicitarRecompensa(clientHokage, clientAkatsuki, statsEquipo)
		case 4:
			// Lógica para ver estadísticas
			mostrarEstadisticas(statsEquipo)
		default:
			log.Println("Opción no válida")
		}
	}
}

/*
	main()
*/
func main() {
	name := os.Getenv("NAME")
	port := os.Getenv("PORT")
	log.Printf("%s is running on port %s", name, port)

	crearEquipoNinja()
}
