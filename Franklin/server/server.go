package server

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
	pb "server.com/franklin/proto"
)

type FranklinServer struct {
	pb.UnimplementedHeistFranklinServiceServer
	mu     sync.Mutex
	turno  int
	dinero int
}

func (s *FranklinServer) RealizarDistraccion(ctx context.Context, req *pb.Solicitud) (*pb.Resultado, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rand.Seed(time.Now().UnixNano())

	turnos := int(200 - req.ProbFranklin)

	for i := 1; i <= turnos; i++ {
		if turnos/2 == i {
			if rand.Intn(100) <= 10 {
				fmt.Printf("Turno %d: Franklin ha sido descubierto en las cercanias del lugar debido a que su perro Chop ha ladrado mucho, abortando mision...\n", i)
				return &pb.Resultado{Res: false}, nil
			}
		}
		fmt.Printf("Turno %d: \n", i)
		time.Sleep(50 * time.Millisecond) // 50ms por turno
	}

	resultado := &pb.Resultado{
		Res: true,
	}

	return resultado, nil
}

func (s *FranklinServer) RealizarAtraco(ctx context.Context, req *pb.SolicitudAtraco) (*pb.Resultado, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.turno = 1
	probExito := req.ProbFranklin

	var conn *amqp091.Connection
	var err error
	rabbitURL := "amqp://guest:guest@rabbitmq:5672/"

	for {
		conn, err = amqp091.Dial(rabbitURL)
		if err != nil {
			log.Println("Esperando a RabbitMQ...", err)
			time.Sleep(2 * time.Second)
			continue
		}
		log.Println("Conectado a RabbitMQ")
		break
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("No se pudo abrir un canal: %v", err)
	}
	defer ch.Close()

	// Declarar cola golpe
	queue, err := ch.QueueDeclare(
		"golpe",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("No se pudo declarar la cola: %v", err)
	}

	msgs, err := ch.Consume(queue.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("No se pudo consumir la cola: %v", err)
	}

	log.Println("Franklin listo para sincronizar turnos...")
	botinExtra := 0
	estrellas := 0

	for msg := range msgs {
		val, err := strconv.Atoi(string(msg.Body))
		if err != nil {
			log.Printf("Valor no válido recibido: %s", msg.Body)
			continue
		}

		switch val {
		case -1:
			// Comenzar a procesar el turno
			time.Sleep(10 * time.Millisecond)
			// Aqui va la logica adicional (si es que hay)
			if estrellas >= 3 {
				botinExtra += 1000
			}

			if estrellas == 5 {
				// Se fracasa el atraco
				err := ch.Publish(
					"", queue.Name, false, false,
					amqp091.Publishing{
						DeliveryMode: amqp091.Persistent,
						ContentType:  "text/plain",
						Body:         []byte(strconv.Itoa(-3)),
					})
				if err != nil {
					log.Printf("Error al enviar -3: %v", err)
				}
				s.dinero = int(req.Botin) + botinExtra
				return &pb.Resultado{Res: false}, nil
			}
			s.turno++
			if 200-probExito < int32(s.turno) {
				// Se termina con exito el atraco
				err := ch.Publish(
					"", queue.Name, false, false,
					amqp091.Publishing{
						DeliveryMode: amqp091.Persistent,
						ContentType:  "text/plain",
						Body:         []byte(strconv.Itoa(-3)),
					})
				if err != nil {
					log.Printf("Error al enviar -3: %v", err)
				}
				s.dinero = int(req.Botin) + botinExtra
				return &pb.Resultado{Res: true}, nil
			}

			// Publicar -2 para avisar que terminó el turno
			err := ch.Publish(
				"", queue.Name, false, false,
				amqp091.Publishing{
					DeliveryMode: amqp091.Persistent,
					ContentType:  "text/plain",
					Body:         []byte(strconv.Itoa(-2)),
				})
			if err != nil {
				log.Printf("Error al enviar -2: %v", err)
			}

		default:
			estrellas = val
		}
	}

	return &pb.Resultado{Res: false}, nil
}
