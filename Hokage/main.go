package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

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
	EquipoCapturador string
}

/*
	AkatsukiList struct
	- Almacena de manera protegida informacion de los Akatsukis detectados.
*/
type AkatsukiList struct {
	Akatsukis map[string]Akatsuki
	mu        sync.Mutex
}

func (al *AkatsukiList) AddAkatsuki(akatsuki Akatsuki) {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.Akatsukis[akatsuki.Nombre] = akatsuki
}

func (al *AkatsukiList) GetAkatsuki(nombre string) (Akatsuki, bool) {
	al.mu.Lock()
	defer al.mu.Unlock()
	akatsuki, exists := al.Akatsukis[nombre]
	return akatsuki, exists
}

func (al *AkatsukiList) ReplaceAkatsuki(nombre string, akatsuki Akatsuki) {
	al.mu.Lock()
	defer al.mu.Unlock()
	if _, exists := al.Akatsukis[nombre]; exists {
		al.Akatsukis[nombre] = akatsuki
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

/*
	listenInformacionAnbu()
- Escucha mensajes en la cola "notificar_akatsuki_Hokage" de RabbitMQ.
- En cada mensaje recibido, deserializa el contenido y muestra la información del Akatsuki detectado.
- Envía una notificación a Hokage y Equipos Ninja con la información del Akatsuki detectado para que puedan actualizar su información local.
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
		
		akatsuki := Akatsuki{
			Id:     int(data.Id),
			Nombre: data.Nombre,
			Ataque: int(data.Ataque),
			Vida:   int(data.Vida),
			Estado: data.Estado,
			EquipoCapturador: data.EquipoCapturador,
		}
		
		if akatsuki.Estado == "CLEAR" {
			s.AkatsukisList.Akatsukis = make(map[string]Akatsuki)
			log.Println("Detectando nuevos Akatsukis, limpiando lista...")
			continue
		}

		log.Println("Mensaje recibido de Anbu")

		akatsukiActual, exists := s.AkatsukisList.GetAkatsuki(akatsuki.Nombre);

		// Si Akatsuki no existe, se asigna recompensa y se registra.
		if !exists {
			log.Printf("Nuevo Akatsuki detectado: %s, asignando recompensa y registrando", akatsuki.Nombre)
			akatsuki.Recompensa = asignarRecompensa(akatsuki)
			s.AkatsukisList.AddAkatsuki(akatsuki)
			continue
		}

		// Si Akatsuki existe, pero fue capturado y reclamada la recompesa, se le asigna nueva recompesa y vuelve a estar Localizado.
		if akatsukiActual.Estado == "Capturado" && akatsukiActual.EquipoCapturador == "" {
			log.Printf("Se a detectado nuevamente al Akatsuki %s... Asignando orden de captura!", akatsuki.Nombre)
			akatsuki.Recompensa = asignarRecompensa(akatsuki)
			s.AkatsukisList.ReplaceAkatsuki(akatsuki.Nombre, akatsuki)
			continue
		}

		// Entonces se mantiene la recompensa y actualiza el estado.
		log.Printf("Actualizando estado del Akatsuki %s a %s", akatsuki.Nombre, akatsuki.Estado)
		akatsuki.Recompensa = akatsukiActual.Recompensa
		s.AkatsukisList.ReplaceAkatsuki(akatsuki.Nombre, akatsuki)
	}
}

/*
	rpc ObtenerListaAkatsukis(ctx context.Context, _ *pbDataHokage.Empty) (*pbDataHokage.ListaAkatsukis, error)
	A.K.A. Mostrar Lista de Akatsukis

	- Entrega la lista de Akatsukis detectados con su información actualizada a quien lo solicite.
*/
func (s *Hokage) ObtenerListaAkatsukis(ctx context.Context, _ *pbDataHokage.Empty) (*pbDataHokage.ListaAkatsukis, error) {
	log.Println("Solicitud de lista de Akatsukis recibida, enviando...")
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
	
	log.Println("Solicitud enviada!")
	return &pbDataHokage.ListaAkatsukis{Akatsukis: data}, nil
}

/*
	rpc ReclamarRecompensa(ctx context.Context, req *pbDataHokage.SolicitudRecompensa) (*pbDataHokage.ConfirmacionPago, error)
	
	- Permite a los Equipos Ninja reclamar la recompensa por capturar a un Akatsuki, siempre y cuando el Akatsuki esté en estado "Capturado" y no haya sido reclamada la recompensa por otro equipo.
*/
func (s *Hokage) ReclamarRecompensa(ctx context.Context, req *pbDataHokage.SolicitudRecompensa) (*pbDataHokage.ConfirmacionPago, error) {
	akatsuki, exists := s.hokageContent.AkatsukisList.GetAkatsuki(req.NombreAkatsuki)
	// No existe el Akatsuki
	if !exists {
		log.Printf("Akatsuki %s no encontrado para reclamar recompensa", req.NombreAkatsuki)
		return &pbDataHokage.ConfirmacionPago{}, fmt.Errorf("Akatsuki %s no encontrado para reclamar recompensa", req.NombreAkatsuki)
	}

	// El Akatsuki no está en estado "Capturado"
	if akatsuki.Estado != "Capturado" {
		log.Printf("Akatsuki %s no está en estado 'Capturado' para reclamar recompensa", req.NombreAkatsuki)
		return &pbDataHokage.ConfirmacionPago{}, fmt.Errorf("Akatsuki %s no está en estado 'Capturado' para reclamar recompensa", req.NombreAkatsuki)
	}

	// El Akatsuki fue capturado pero la recompensa ya fue reclamada
	if akatsuki.Estado == "Capturado" && akatsuki.EquipoCapturador == "" {
		log.Printf("Akatsuki %s fue capturado pero la recompensa ya fue reclamada, no puede ser reclamada por el equipo %s", req.NombreAkatsuki, req.NombreEquipo)
		return &pbDataHokage.ConfirmacionPago{}, fmt.Errorf("Akatsuki %s fue capturado pero la recompensa ya fue reclamada, no puede ser reclamada", req.NombreAkatsuki)
	}

	// El Akatsuki fue capturado por otro equipo
	if akatsuki.EquipoCapturador != "" && akatsuki.EquipoCapturador != req.NombreEquipo {
		log.Printf("Akatsuki %s ya fue capturado por el equipo %s, no puede ser reclamado por el equipo %s", req.NombreAkatsuki, akatsuki.EquipoCapturador, req.NombreEquipo)
		return &pbDataHokage.ConfirmacionPago{}, fmt.Errorf("Akatsuki %s ya fue capturado por el equipo %s, no puede ser reclamado", req.NombreAkatsuki, akatsuki.EquipoCapturador)
	}

	log.Printf("Recompensa de %d ryo reclamada por equipo %s por capturar a %s", akatsuki.Recompensa, req.NombreEquipo, akatsuki.Nombre)

	data := &pbDataHokage.ConfirmacionPago{
		Mensaje: "Recompensa reclamada exitosamente",
		RyoPagados: int32(akatsuki.Recompensa),
	}

	// Actualizar estado del Akatsuki a "Capturado" y desasociar el equipo capturador
	akatsuki.EquipoCapturador = ""
	s.hokageContent.AkatsukisList.ReplaceAkatsuki(akatsuki.Nombre, akatsuki)

	return data, nil
}

/*
	serverBackground()

	- Inicializa el servidor gRPC de Hokage y la tarea de escuchar mensajes de RabbitMQ en segundo plano.
*/
func serverBackground() {
	hokageContent := &HokageContent{
		AkatsukisList: &AkatsukiList{
			Akatsukis: make(map[string]Akatsuki),
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

/*
	asignarRecompensa(Akatsuki Akatsuki) int
	- Calcula la recompensa para un Akatsuki basado en sus atributos de ataque y vida.
*/
func asignarRecompensa(Akatsuki Akatsuki) int {
	// Recompensa base de 1000, más 10 por cada punto de ataque y vida
	return 1000 + (Akatsuki.Ataque * 10) + (Akatsuki.Vida * 10)
}


/*
	mian()
*/
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
