package main

import (
	"context"
	"log"
	"net"
	"os"
	"sync"
	"time"
	"fmt"

	pbDataAkatsuki "Hokage/proto/DataAkatsuki"
	pbDataHokage "Hokage/proto/DataHokage"

	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type Akatsuki struct {
	Id         int
	Nombre     string
	Ataque     int
	Vida       int
	Estado     string // "Localizado", "En Combate", "Capturado"
	Recompensa int    // Recompensa asociada
}

type AkatsukiList struct {
	Akatsukis map[int]Akatsuki
	mu        sync.Mutex
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

func (al *AkatsukiList) ReplaceEstado(id int, estado string) {
	al.mu.Lock()
	defer al.mu.Unlock()
	if akatsuki, exists := al.Akatsukis[id]; exists {
		akatsuki.Estado = estado
		al.Akatsukis[id] = akatsuki
	}
}

func (al *AkatsukiList) GetAllAkatsukis() []Akatsuki {
	al.mu.Lock()
	defer al.mu.Unlock()
	akatsukis := make([]Akatsuki, 0, len(al.Akatsukis))
	for _, akatsuki := range al.Akatsukis {
		akatsukis = append(akatsukis, akatsuki)
	}
	return akatsukis
}

type HokageContent struct {
	AkatsukisList *AkatsukiList
}

type Hokage struct {
	pbDataHokage.UnimplementedHokageServer
	hokageContent *HokageContent
}

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

/*
	Escucha información de Anbu
*/
func (s *HokageContent) listenInformacionAnbu() {
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

	msgs, err := ch.Consume(
		"notificar_akatsuki_Hokage",
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
		log.Printf("Mensaje recibido de Anbu: %d", data.Id)

		akatsuki := Akatsuki{
			Id:     int(data.Id),
			Nombre: data.Nombre,
			Ataque: int(data.Ataque),
			Vida:   int(data.Vida),
			Estado: data.Estado,
		}

		if _, exists := s.AkatsukisList.GetAkatsuki(akatsuki.Id); exists {
			// Si el Akatsuki ya existe, actualizamos su estado
			s.AkatsukisList.ReplaceEstado(akatsuki.Id, akatsuki.Estado)

		} else {
			// Si el Akatsuki no existe, lo agregamos a la lista y le asignamos una recompensa
			akatsuki.Recompensa = asignarRecompensa(akatsuki)
			s.AkatsukisList.AddAkatsuki(akatsuki)
		}
	}
}

/*
	A.K.A. Mostrar Lista de Akatsukis
*/
func (s *Hokage) ObtenerListaAkatsukis(ctx context.Context, _ *pbDataHokage.Empty) (*pbDataHokage.ListaAkatsukis, error) {
	akatsukiList := s.hokageContent.AkatsukisList.GetAllAkatsukis()
	data := make([]*pbDataHokage.AkatsukiInfo, 0, len(akatsukiList))
	for _, akatsuki := range akatsukiList {
		data = append(data, &pbDataHokage.AkatsukiInfo{
			Id:        int32(akatsuki.Id),
			Nombre:    akatsuki.Nombre,
			Ataque:    int32(akatsuki.Ataque),
			Vida:      int32(akatsuki.Vida),
			Estado:    akatsuki.Estado,
			Recompensa: int32(akatsuki.Recompensa),
		})
	}
	
	return &pbDataHokage.ListaAkatsukis{Akatsukis: data}, nil
}

func (s *Hokage) ReclamarRecompensa(ctx context.Context, req *pbDataHokage.SolicitudRecompensa) (*pbDataHokage.ConfirmacionPago, error) {

	return &pbDataHokage.ConfirmacionPago{}, nil
}

func serverBackground() {
	hokageContent := &HokageContent{
		AkatsukisList: &AkatsukiList{
			Akatsukis: make(map[int]Akatsuki),
		},
	}

	go hokageContent.listenInformacionAnbu()

	listener, err := net.Listen("tcp", ":"+os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pbDataHokage.RegisterHokageServer(grpcServer, &Hokage{hokageContent: hokageContent})
	log.Printf("Hokage server is listening on port %s", os.Getenv("PORT"))
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func asignarRecompensa(Akatsuki Akatsuki) int {
	// Recompensa base de 1000, más 10 por cada punto de ataque y vida
	return 1000 + (Akatsuki.Ataque * 10) + (Akatsuki.Vida * 10)
}

func main() {
	// Creacion de instancia de servidor
	name := os.Getenv("NAME")
	port := os.Getenv("PORT")
	log.Printf("%s is running on port %s", name, port)

	go serverBackground()

	forever := make(chan bool)
	log.Printf("Waiting for messages. To exit press CTRL+C")
	<-forever
}
