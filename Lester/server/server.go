package server

import (
	"context"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
	pb "server.com/lester/proto"
)

type LesterServer struct {
	pb.UnimplementedHeistServiceServer
	mu         sync.Mutex
	rechazos   int
	riesgo     int
	estrellas  int
	turno      int
	rabbitConn *amqp091.Connection
	rabbitCh   *amqp091.Channel
}

func NewLesterServer() *LesterServer {
	s := &LesterServer{
		rechazos:  0,
		riesgo:    0,
		estrellas: 0,
		turno:     1,
	}

	go func() {
		for {
			// Intentar conectarse a RabbitMQ
			log.Println("Haciendo conexión con RabbitMQ.: ")
			url := os.Getenv("RABBITMQ_URL")
			conn, err := amqp091.Dial(url)
			if err != nil {
				log.Println("Esperando a RabbitMQ...: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}
			log.Println("Creando canal.")
			ch, err := conn.Channel()
			if err != nil {
				log.Println("Error creando canal:", err)
				conn.Close()
				time.Sleep(2 * time.Second)
				continue
			}

			log.Println("Creando canal.")
			// Declarar la cola antes de consumir
			_, err = ch.QueueDeclare(
				"golpe",
				true,  // durable
				false, // autoDelete
				false, // exclusive
				false, // noWait
				nil,   // args
			)
			if err != nil {
				log.Println("Error declarando cola:", err)
				ch.Close()
				conn.Close()
				time.Sleep(2 * time.Second)
				continue
			}

			log.Println("Guardando conexion.")
			// Guardar conexión y canal en el servidor
			s.rabbitConn = conn
			s.rabbitCh = ch

			// Arrancar consumidor de la cola
			go s.consumeGolpe()
			log.Println("Conectado a RabbitMQ y consumidor iniciado")
			break
		}
	}()

	return s
}

func (s *LesterServer) SolicitarOferta(ctx context.Context, req *pb.Solicitud) (*pb.Oferta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Peticion de oferta recibida")

	rand.Seed(time.Now().UnixNano())

	// Probabilidad oferta
	if rand.Intn(100) >= 90 {
		return &pb.Oferta{Disponible: false}, nil
	}

	// Espera de 10 segundos
	if s.rechazos >= 3 {
		log.Printf("Se rechazaron 3 ofertas simultaneamente, así que se enojó D:< ")
		time.Sleep(10 * time.Second)
		s.rechazos = 0
	}

	oferta := &pb.Oferta{
		Botin:          int32(rand.Intn(1_000_000) + 100_000),
		ProbFranklin:   int32(rand.Intn(101)),
		ProbTrevor:     int32(rand.Intn(101)),
		RiesgoPolicial: int32(rand.Intn(101)),
		Disponible:     true,
	}
	return oferta, nil
}

func (s *LesterServer) AceptarOferta(ctx context.Context, dec *pb.Decision) (*pb.Respuesta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if dec.Aceptada {
		s.rechazos = 0
		return &pb.Respuesta{Mensaje: "Atraco aceptado por Michael"}, nil
	}
	s.rechazos++
	return &pb.Respuesta{Mensaje: "Oferta rechazada"}, nil
}

// Manda a la cola RabbitMQ el valor ingresado
func (s *LesterServer) publishGolpe(valor int) (*pb.Empty, error) {
	for s.rabbitCh == nil {
		log.Println("RabbitMQ no está listo, esperando...")
		time.Sleep(500 * time.Millisecond)
	}
	err := s.rabbitCh.Publish(
		"", "golpe", false, false,
		amqp091.Publishing{
			DeliveryMode: amqp091.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(strconv.Itoa(valor)),
		})
	if err != nil {
		log.Printf("Error publicando en la cola: %v", err)
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (s *LesterServer) EstudiarGolpe(ctx context.Context, riesgo *pb.Riesgo) (*pb.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.riesgo = int(riesgo.Risk)
	return s.publishGolpe(-1)
}

func (s *LesterServer) consumeGolpe() {
	msgs, err := s.rabbitCh.Consume("golpe", "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Error al consumir la cola: %v", err)
	}

	for msg := range msgs {
		val, err := strconv.Atoi(string(msg.Body))
		if err != nil {
			log.Printf("Valor no válido recibido: %s", msg.Body)
			continue
		}

		switch val {
		case -2:
			log.Println("Turno finalizado, enviando nuevo turno...")
			if s.turno%(100-s.riesgo) == 0 {
				s.estrellas = s.estrellas + 1
				if _, err := s.publishGolpe(s.estrellas); err != nil {
					log.Printf("Error al publicar nueva estrella: %v", err)
				}
			}
			s.turno++
			if _, err := s.publishGolpe(-1); err != nil {
				log.Printf("Error al publicar nuevo turno: %v", err)
			}

		case -3:
			log.Println("Lester ya no necesita seguir revisando el nivel de busqueda.")
			return
		}
	}
}
