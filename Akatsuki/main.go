package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"

	pbCombateAkatsuki "Akatsuki/proto/CombateAkatsuki"
	pbDataAkatsuki "Akatsuki/proto/DataAkatsuki"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
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
		time.Sleep(5 * time.Second)
	}
	return nil, fmt.Errorf("could not connect to RabbitMQ after %d attempts", maxAttempts)
}

type Akatsuki struct {
	Id         int
	Nombre     string
	Ataque     int
	Vida       int
	Estado     string // "Localizado", "En Combate", "Capturado"
	EquipoCapturador string
}

type akatsukiData struct {
	mu sync.Mutex
	enemigos map[string]Akatsuki
}

/*
	LogSystem
	- Estructura para almacenar los logs de los combates.
	- Se utiliza un mutex para asegurar que el acceso a los logs sea seguro en concurrencia.
*/
type logSystem struct {
	mu sync.Mutex
	logs []string
}

func (ls *logSystem) AddLog(logEntry string) {
	ls.logs = append(ls.logs, logEntry)
}

func (ls *logSystem) Lock() {
	ls.mu.Lock()
}

func (ls *logSystem) Unlock() {
	ls.mu.Unlock()
}

func (ls *logSystem) PrintLogs() {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	for _, logEntry := range ls.logs {
		fmt.Print(logEntry)
	}
	ls.logs = []string{}
}

type combateAkatsukiServer struct {
	pbCombateAkatsuki.UnimplementedCombateAkatsukiServer
	logSystem *logSystem
	akatsukiData *akatsukiData
	conn *amqp091.Connection
}

/*
	localizarAkatsuki(conn *amqp091.Connection)
- Genera un nuevo Akatsuki cada cierto tiempo con atributos aleatorios y lo agrega a la lista de enemigos.
- Envía un mensaje a RabbitMQ en localizar_akatsuki cada vez que se genera un nuevo Akatsuki otros nodos puedan actualizar su información local.
*/
func (s *akatsukiData) localizarAkatsuki(conn *amqp091.Connection) {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error al abrir canal RabbitMQ: %v", err)
	}
	defer ch.Close()

	data := &pbDataAkatsuki.DataAkatsukiRequest{
	Id:     int32(-1),
	Nombre: "CLEAR",
	Ataque: int32(-1),
	Vida:   int32(-1),
	Estado: "CLEAR",
	EquipoCapturador: "",
	}
	body, err := proto.Marshal(data)
	if err != nil {
		log.Fatalf("Error al serializar datos: %v", err)
	}

	err = ch.Publish(
		"",     			  // exchange
		"localizar_akatsuki", // routing key
		false,  			  // mandatory
		false,  			  // immediate
		amqp091.Publishing{
			ContentType: "application/protobuf",
			Body:        body,
	})
	if err != nil {
		log.Fatalf("Error al publicar mensaje: %v", err)
	}

	nombres := []string{"Konan", "Nagato", "Yahiko", "Kyusuke", "Kie", "Daibutsu", "Obito Uchiha", "Zetsu Blanco", "Zetsu Negro", "Kisame Hoshigaki", "Konan", "Nagato", "Itachi Uchiha", "Deidara", "Orochimaru", "Juzo Biwa", "Kakuzu", "Hidan", "Sasori"}
	apariciones := make(map[string]int, 0)

	for {
		// Espera un tiempo aleatorio entre 0 y 20 segundos antes de generar un nuevo Akatsuki
		time.Sleep(time.Duration(rand.Intn(20)) * time.Second)

		// Generar Akatuski
		var newNombre string
		var id int
		for {
			s.mu.Lock()
			id = rand.Intn(len(nombres))
			newNombre = nombres[id]
			// Akatsuki nuevo
			if _, exists := s.enemigos[newNombre]; !exists {
				apariciones[newNombre]++
				s.mu.Unlock()
				break
			}
			
			// Si el Akatsuki ya existe pero está capturado sin equipo capturador, se puede generar uno nuevo con el mismo nombre
			if s.enemigos[newNombre].Estado == "Capturado" && s.enemigos[newNombre].EquipoCapturador == "" {
				apariciones[newNombre]++
				s.mu.Unlock()
				break;
			}
			s.mu.Unlock()
			time.Sleep(1 * time.Second)
		}

		nuevo := Akatsuki{
			Id:         id,
			Nombre:     newNombre,
			Ataque:     40 + rand.Intn(31) + apariciones[newNombre] * 5,
			Vida:       65 + rand.Intn(71) + apariciones[newNombre] * 5,
			Estado:     "Localizado",
		}

		s.mu.Lock()
		s.enemigos[nuevo.Nombre] = nuevo
		s.mu.Unlock()

		data := &pbDataAkatsuki.DataAkatsukiRequest{
			Id:     int32(nuevo.Id),
			Nombre: nuevo.Nombre,
			Ataque: int32(nuevo.Ataque),
			Vida:   int32(nuevo.Vida),
			Estado: nuevo.Estado,
			EquipoCapturador: "",
		}
		body, err := proto.Marshal(data)
		if err != nil {
			log.Fatalf("Error al serializar datos: %v", err)
		}

		err = ch.Publish(
			"",     			  // exchange
			"localizar_akatsuki", // routing key
			false,  			  // mandatory
			false,  			  // immediate
			amqp091.Publishing{
				ContentType: "application/protobuf",
				Body:        body,
			})
		if err != nil {
			log.Fatalf("Error al publicar mensaje: %v", err)
		}
	}
}

/*
	actualizarEstado(nombre string, estado string, equipoCapturador string, conn *amqp091.Connection)
	- Actualiza el estado y equipo capturador de un Akatsuki específico.
	- Envía un mensaje a RabbitMQ en localizar_akatsuki, con la información actualizada del Akatsuki para que otros nodos puedan actualizar su información local.
*/
func (s *akatsukiData) actualizarEstado(nombre string, estado string, equipoCapturador string, conn *amqp091.Connection) {
	s.mu.Lock()
	akatsuki, exists := s.enemigos[nombre]
	if !exists {
		s.mu.Unlock()
		log.Printf("Akatsuki %s no encontrado para actualizar estado", nombre)
		return
	}
	akatsuki.Estado = estado
	akatsuki.EquipoCapturador = equipoCapturador
	s.enemigos[nombre] = akatsuki
	s.mu.Unlock()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error al abrir canal RabbitMQ: %v", err)
	}
	defer ch.Close()

	data := &pbDataAkatsuki.DataAkatsukiRequest{
		Id:     int32(akatsuki.Id),
		Nombre: akatsuki.Nombre,
		Ataque: int32(akatsuki.Ataque),
		Vida:   int32(akatsuki.Vida),
		Estado: akatsuki.Estado,
		EquipoCapturador: akatsuki.EquipoCapturador,
	}
	body, err := proto.Marshal(data)
	if err != nil {
		log.Fatalf("Error al serializar datos: %v", err)
	}

	err = ch.Publish(
		"",     			  // exchange
		"localizar_akatsuki", // routing key
		false,  			  // mandatory
		false,  			  // immediate
		amqp091.Publishing{
			ContentType: "application/protobuf",
			Body:        body,
		})
	if err != nil {
		log.Fatalf("Error al publicar mensaje: %v", err)
	}
}

type equipoNinja struct {
	Nombre string
	Atk int
	Vid int
}

/*
	rpc IniciarCombate(IniciarCombateRequest) returns (IniciarCombateResult)
	- Inicializa un combate entre un equipo ninja y un Akatsuki localizado.
	- Retorna un resultado indicando si el equipo ninja logró capturar al Akatsuki o si este escapó o derrotó al equipo ninja.
*/
func (s *combateAkatsukiServer) IniciarCombate(ctx context.Context, req *pbCombateAkatsuki.IniciarCombateRequest) (*pbCombateAkatsuki.IniciarCombateResult, error) {
	var ninja, rival int;
	var turnoNinja bool;
	// Extrae la data de los participantes
	equipoNinja := equipoNinja{
		Nombre: req.Equipo.NombreEquipo,
		Atk: int(req.Equipo.Ataque),
		Vid: int(req.Equipo.Vida),
	}

	// Bloquear acceso a Data de Akatsuki para verificar estado.
	s.akatsukiData.mu.Lock()
	var enemigo Akatsuki

	if _, exists := s.akatsukiData.enemigos[req.NombreObjetivo]; !exists {
		fmt.Printf("Akatsuki %s se encuentra oculto!\n", req.NombreObjetivo)
		s.akatsukiData.mu.Unlock()
		return &pbCombateAkatsuki.IniciarCombateResult{Success: false}, fmt.Errorf("Akatsuki no Localizado")
	}

	if s.akatsukiData.enemigos[req.NombreObjetivo].Estado != "Localizado" {
		s.akatsukiData.mu.Unlock()
		return &pbCombateAkatsuki.IniciarCombateResult{Success: false}, fmt.Errorf("Akatsuki Capturado o en Combate")
	}

	enemigo = s.akatsukiData.enemigos[req.NombreObjetivo]
	s.akatsukiData.mu.Unlock()
	s.akatsukiData.actualizarEstado(enemigo.Nombre, "En Combate", "", s.conn)

	// Se calcula cual de los 2 equipos ataca primero
	ninja = rand.Intn(100)
	rival = rand.Intn(100)

	if ninja > rival {
		turnoNinja = true
	} else {
		turnoNinja = false
	}

	// Iniciar Combate, Bloquea salida de logs.
	s.logSystem.Lock()
	s.logSystem.AddLog("=========================================================\n")
	s.logSystem.AddLog("==                LOGS DEL COMBATE                     ==\n")
	s.logSystem.AddLog("=========================================================\n")
	s.logSystem.AddLog("\nComienza Combate\n")
	s.logSystem.AddLog("Equipo " + equipoNinja.Nombre + " VS " + enemigo.Nombre + "\n")

	result := pbCombateAkatsuki.IniciarCombateResult{Success: true}

	for {
		var mult float32
		var miss float32
		mult = 1.5 * rand.Float32() + 0.5
		miss = rand.Float32()

		if turnoNinja {
			// El equipo ninja ataca
			s.logSystem.AddLog("\nEquipo " + equipoNinja.Nombre + " ataca\n")
			ataque := int(float32(equipoNinja.Atk) * mult)

			// 30% de probabilidad de fallar el ataque
			if miss <= 0.3 {
				s.logSystem.AddLog("¡El ataque falló!\n")
			} else {
				enemigo.Vida -= ataque
				s.logSystem.AddLog("" + enemigo.Nombre + " recibió " + strconv.Itoa(ataque) + " de daño. Vida restante: " + strconv.Itoa(enemigo.Vida) + "\n")
			}

		} else {
			// El Akatsuki ataca
			// Inteno de Escape
			var escape float32
			escape = rand.Float32()
			if escape <= 0.1 {
				s.logSystem.AddLog("\n" + enemigo.Nombre + " ha escapado del combate!\n")
				result.Success = false
				s.akatsukiData.actualizarEstado(enemigo.Nombre, "Localizado", "", s.conn)
				break
			}

			s.logSystem.AddLog("\n" + enemigo.Nombre + " atacó\n")
			ataque := int(float32(enemigo.Ataque) * mult)

			// 30% de probabilidad de fallar el ataque
			if miss <= 0.3 {
				s.logSystem.AddLog("¡El ataque falló!\n")
			} else {
				equipoNinja.Vid -= ataque
				s.logSystem.AddLog("" + equipoNinja.Nombre + " recibió " + strconv.Itoa(ataque) + " de daño. Vida restante: " + strconv.Itoa(equipoNinja.Vid) + "\n")
			}
		}

		// Verificar si alguno de los 2 equipos ha sido derrotado
		if equipoNinja.Vid < 0  {
			s.logSystem.AddLog("\n" + enemigo.Nombre + " Derrotó al Equipo " + equipoNinja.Nombre + ", ha escapado\n")
			result.Success = false
			s.akatsukiData.actualizarEstado(enemigo.Nombre, "Localizado", "", s.conn)
			break
		}

		if enemigo.Vida < 0 {
			s.logSystem.AddLog("\nEquipo " + equipoNinja.Nombre + " Derrotó a " + enemigo.Nombre + ", ha sido capturado\n")
			result.Success = true
			s.akatsukiData.actualizarEstado(enemigo.Nombre, "Capturado", equipoNinja.Nombre, s.conn)
			break
		}

		turnoNinja = !turnoNinja
		time.Sleep(time.Duration(rand.Float32() * 2.0) * time.Second)
	}

	s.logSystem.AddLog("\nCombate finalizado\n")
	s.logSystem.AddLog("=========================================================\n")

	s.logSystem.Unlock()
	s.logSystem.PrintLogs()
	return &result, nil
}

/*
	rpc LiberarAkatsuki(LiberarAkatsukiRequest) returns (LiberarAkatsukiResult)
	- Permite liberar un Akatsuki que se encuentra capturado, cambiando su estado a "Localizado" y su equipo capturador a "".
	- Solo se puede liberar un Akatsuki que se encuentra en estado "Capturado".
*/
func (s *combateAkatsukiServer) LiberarAkatsuki(ctx context.Context, req *pbCombateAkatsuki.LiberarAkatsukiRequest) (*pbCombateAkatsuki.LiberarAkatsukiResult, error) {
	s.akatsukiData.mu.Lock()
	akatsuki, exists := s.akatsukiData.enemigos[req.NombreAkatsuki]
	if !exists {
		s.akatsukiData.mu.Unlock()
		log.Printf("Akatsuki %s no encontrado para liberar", req.NombreAkatsuki)
		return &pbCombateAkatsuki.LiberarAkatsukiResult{Success: false}, fmt.Errorf("Akatsuki no encontrado")
	}
	if akatsuki.Estado != "Capturado" {
		s.akatsukiData.mu.Unlock()
		log.Printf("Akatsuki %s no está capturado, no se puede liberar", req.NombreAkatsuki)
		return &pbCombateAkatsuki.LiberarAkatsukiResult{Success: false}, fmt.Errorf("Akatsuki no está capturado")
	}

	akatsuki.EquipoCapturador = ""
	s.akatsukiData.enemigos[req.NombreAkatsuki] = akatsuki
	
	s.akatsukiData.mu.Unlock()
	return &pbCombateAkatsuki.LiberarAkatsukiResult{Success: true}, nil
}

/*
	serverBackground()
	- Realiza la conexión a RabbitMQ
	- Inicializa el servicio gRPC
	- Inicializa la estructura de datos de los Akatsukis y logSystem
*/
func serverBackground() {
	server := &akatsukiData{
		enemigos: make(map[string]Akatsuki),
	}

	logSystem := &logSystem{
		logs: make([]string, 0),
	}

	conn, err := dialRabbitMQ()
	if err != nil {
		log.Fatalf("Error al conectar con RabbitMQ: %v", err)
	}
	defer conn.Close()

	listener, err := net.Listen("tcp", ":"+os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go server.localizarAkatsuki(conn)

	grpcServer := grpc.NewServer()
	pbCombateAkatsuki.RegisterCombateAkatsukiServer(grpcServer, &combateAkatsukiServer{akatsukiData: server, logSystem: logSystem, conn: conn})
	log.Printf("Hokage server is listening on port %s", os.Getenv("PORT"))
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

/*
	Main()
	- Inicia la hebra del nodo.
*/
func main() {
	name := os.Getenv("NAME")
	port := os.Getenv("PORT")
	log.Printf("%s is running on port %s", name, port)

	go serverBackground()

	forever := make(chan bool)
	log.Printf("To exit press CTRL+C")
	<-forever
}
