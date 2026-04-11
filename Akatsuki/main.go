package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	pb "Akatsuki/proto/message"

	"google.golang.org/grpc"
)

type Enemigo struct {
	Nombre     string
	Ataque     int
	Vida       int
	Estado     string // "Localizado", "En Combate", "Capturado" [cite: 91]
	Recompensa string
	Disponible bool
}

type akatsukiServer struct {
	pb.UnimplementedAkatsukiServer
	mu       sync.Mutex
	enemigos []Enemigo
}

// Tarea: Generar Akatsuki de forma periódica
func (s *akatsukiServer) generarEnemigos() {
	nombres := []string{"Itachi", "Kisame", "Nagato", "Sasori", "Deidara", "Hidan", "Kakuzu"}
	for {
		time.Sleep(20 * time.Second) // Intervalo de aparición
		s.mu.Lock()
		nuevo := Enemigo{
			Nombre:     nombres[rand.Intn(len(nombres))],
			Ataque:     55 + rand.Intn(31), // 70 +- 30
			Vida:       75 + rand.Intn(76), // 150 +- 75
			Estado:     "Localizado",
			Recompensa: fmt.Sprintf("%dM", 5+rand.Intn(11)),
		}
		s.enemigos = append(s.enemigos, nuevo)
		fmt.Printf("[SISTEMA] Nuevo Akatsuki detectado: %s\n", nuevo.Nombre)
		s.mu.Unlock()
		// Se sube la cosa a la cola Rabbit

	}
}

// Comenzar Combate y Simulación
func (s *akatsukiServer) IniciarCombate(ctx context.Context, req *pb.DatosEquipo) (*pb.EstadoCombate, error) {

}

func main() {
	// Inicializar servidor gRPC
	lis, err := net.Listen("tcp", ":50051") // Puerto asignado en MV3
	if err != nil {
		log.Fatalf("Fallo al escuchar: %v", err)
	}

	s := grpc.NewServer()
	serverInstance := &akatsukiServer{
		enemigos: []Enemigo{
			{Nombre: "Itachi", Ataque: 85, Vida: 180, Estado: "Localizado", Recompensa: "8M"},
		},
	}

	pb.RegisterAkatsukiServer(s, serverInstance)

	conn, _ := amqp.Dial("amqp://guest:guest@anbu-vm:5672/")
	ch, _ := conn.Channel()

	// Declarar la cola por seguridad
	ch.QueueDeclare("lista_akatsukis", false, false, false, false, nil)

	// Goroutine para generar enemigos dinámicamente
	go serverInstance.generarEnemigos()

	fmt.Println("Servidor Akatsuki (MV3) ejecutándose en puerto 50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Fallo al servir: %v", err)
	}
}
