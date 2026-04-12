package main

import (
	// "context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"

	pb "Akatsuki/proto/DataAkatsuki"

	// "google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func dialRabbitMQ() (*amqp091.Connection, error) {
	log.Printf("Connecting to RabbitMQ at %s", "amqp://guest:guest@" + os.Getenv("RABBITMQ-IP") + "/")
	conn, err := amqp091.Dial("amqp://guest:guest@" + os.Getenv("RABBITMQ-IP") + "/")
	if err != nil {
		return nil, err
	}
	return conn, nil
}

type Enemigo struct {
	Nombre     string
	Ataque     int
	Vida       int
	Estado     string // "Localizado", "En Combate", "Capturado" [cite: 91]
	// Recompensa string
	Disponible bool
}

type akatsukiServer struct {
	mu       sync.Mutex
	enemigos []Enemigo
}

// Tarea: Generar Akatsuki de forma periódica
func (s *akatsukiServer) generarEnemigos() {
	nombres := []string{"Itachi", "Kisame", "Nagato", "Sasori", "Deidara", "Hidan", "Kakuzu"}
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
		nuevo := Enemigo{
			Nombre:     nombres[rand.Intn(len(nombres))],
			Ataque:     55 + rand.Intn(31), // 70 +- 30
			Vida:       75 + rand.Intn(76), // 150 +- 75
			Estado:     "Localizado",
		}
		s.enemigos = append(s.enemigos, nuevo)
		fmt.Printf("[SISTEMA] Nuevo Akatsuki detectado: %s\n", nuevo.Nombre)
		s.mu.Unlock()
		// Se sube la cosa a la cola Rabbit

		data := &pb.DataAkatsukiRequest{
			Id:     int32(len(s.enemigos)-1), // ID basado en la longitud actual
			Nombre:   nuevo.Nombre,
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


// Comenzar Combate y Simulación
// func (s *akatsukiServer) IniciarCombate(ctx context.Context, req *pb.DatosEquipo) (*pb.EstadoCombate, error) {

// }

func main() {
	name := os.Getenv("NAME")
	port := os.Getenv("PORT")
	log.Printf("%s is running on port %s", name, port)
	// Inicializar servidor gRPC
	// lis, err := net.Listen("tcp", ":50051") // Puerto asignado en MV3
	// if err != nil {
	// 	log.Fatalf("Fallo al escuchar: %v", err)
	// }

	// s := grpc.NewServer()
	serverInstance := &akatsukiServer{
		enemigos: []Enemigo{
			{Nombre: "Itachi", Ataque: 85, Vida: 180, Estado: "Localizado", Disponible: true},
		},
	}

	// // pb.RegisterAkatsukiServer(s, serverInstance)

	// conn, _ := amqp.Dial("amqp://guest:guest@anbu-vm:5672/")
	// ch, _ := conn.Channel()

	// // Declarar la cola por seguridad
	// ch.QueueDeclare("lista_akatsukis", false, false, false, false, nil)

	// Goroutine para generar enemigos dinámicamente
	go serverInstance.generarEnemigos()

	forever := make(chan bool)
	log.Printf("Waiting for messages. To exit press CTRL+C")
	<-forever
}
