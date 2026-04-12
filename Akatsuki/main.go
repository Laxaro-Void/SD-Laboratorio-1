package main

import (
	"context"
	"net"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"

	pbDataAkatsuki "Akatsuki/proto/DataAkatsuki"
	pbCombateAkatsuki "Akatsuki/proto/CombateAkatsuki"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
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

type Akatsuki struct {
	Nombre     string
	Ataque     int
	Vida       int
	Estado     string // "Localizado", "En Combate", "Capturado"
}

type akatsukiData struct {
	mu sync.Mutex
	enemigos []Akatsuki
}

type akatsukiServer struct {
	pbCombateAkatsuki.UnimplementedCombateAkatsukiServer
}

// Generar Akatsuki de forma periódica
func (s *akatsukiData) localizarAkatsuki() {
	nombres := []string{"Konan", "Nagato", "Yahiko", "Kyusuke", "Kie", "Daibutsu", "Obito Uchiha", "Zetsu Blanco", "Zetsu Negro", "Kisame Hoshigaki", "Konan", "Nagato", "Itachi Uchiha", "Deidara", "Orochimaru", "Juzo Biwa", "Kakuzu", "Hidan", "Sasori"}
	conn, err := dialRabbitMQ()
	if err != nil {
		log.Fatalf("Error al conectar con RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error al abrir canal RabbitMQ: %v", err)
	}
	defer ch.Close()
	
	for {
		time.Sleep(20 * time.Second) // Intervalo de aparición
		s.mu.Lock()
		nuevo := Akatsuki{
			Nombre:     nombres[rand.Intn(len(nombres))],
			Ataque:     55 + rand.Intn(31), // 70 +- 30
			Vida:       75 + rand.Intn(76), // 150 +- 75
			Estado:     "Localizado",
		}
		s.enemigos = append(s.enemigos, nuevo)
		fmt.Printf("[SISTEMA] Nuevo Akatsuki detectado: %s\n", nuevo.Nombre)
		s.mu.Unlock()

		data := &pbDataAkatsuki.DataAkatsukiRequest{
			Id:     int32(len(s.enemigos)-1),
			Nombre: nuevo.Nombre,
			Ataque: int32(nuevo.Ataque),
			Vida:   int32(nuevo.Vida),
			Estado: nuevo.Estado,
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

func (s *akatsukiData) actualizarEstado(id int, estado string) {

}

// Comenzar Combate y Simulación
func (s *akatsukiServer) IniciarCombate(ctx context.Context, req *pbCombateAkatsuki.IniciarCombateRequest) (*pbCombateAkatsuki.IniciarCombateResult, error) {
	var ninja, rival int;
	var turnoNinja bool;

	// Se calcula cual de los 2 equipos ataca primero
	for ninja == rival {
		ninja = rand.Intn(100)
		rival = rand.Intn(100)
	}

	if ninja > rival {
		turnoNinja = true
	} else {
		turnoNinja = false
	}

	fmt.Println("=========================================================")
	fmt.Println("==                LOGS DEL COMBATE                     ==")
	fmt.Println("=========================================================")
	fmt.Println("\nComienza Combate")

	if turnoNinja {
		
	}

	return &pbCombateAkatsuki.IniciarCombateResult{Success: true}, nil
}

func serverBackground() {
	server := &akatsukiData{
		enemigos: make([]Akatsuki, 0),
	}

	go server.localizarAkatsuki()

	listener, err := net.Listen("tcp", ":"+os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pbCombateAkatsuki.RegisterCombateAkatsukiServer(grpcServer, &akatsukiServer{})
	log.Printf("Hokage server is listening on port %s", os.Getenv("PORT"))
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func main() {
	name := os.Getenv("NAME")
	port := os.Getenv("PORT")
	log.Printf("%s is running on port %s", name, port)

	go serverBackground()

	forever := make(chan bool)
	log.Printf("Waiting for messages. To exit press CTRL+C")
	<-forever
}
