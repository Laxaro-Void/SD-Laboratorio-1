package main

import (
	"fmt"
	"log"
	"os"
	"time"

	pb "Anbu/proto/DataAkatsuki"
	"github.com/rabbitmq/amqp091-go"
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
		time.Sleep(5 * time.Second) // Wait before retrying
	}
	return nil, fmt.Errorf("could not connect to RabbitMQ after %d attempts", maxAttempts)
}

/*
	initialize(conn *amqp091.Connection)
- Declara los intercambios y colas necesarias en RabbitMQ para la comunicación entre Anbu, Hokage y Equipos Ninja.
- Se asegura de que las colas "notificar_akatsuki_Hokage" y "localizar_akatsuki" estén declaradas y listas para su uso.
*/
func initialize(conn *amqp091.Connection) {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
		return
	}
	defer ch.Close()

	// Intercambiador, Equipos Ninjas se subscriben a este punto con sus colas propias.
	err = ch.ExchangeDeclare(
		"EquiposNinjaBrodcast", // name
		"fanout",               // type
		true,                   // durable
		false,                  // auto-deleted
		false,                  // internal
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Colas para comunicación directa entre Anbu y Hokage
	q, _ := ch.QueueDeclare(
		"notificar_akatsuki_Hokage", // name
		false,  			  		 // durable
		false,  			  		 // delete when unused
		false,  			  		 // exclusive
		false,  			  		 // no-wait
		nil,    			  		 // arguments
	)

	log.Printf("Declared queue: %s", q.Name)

	// Cola para comunicación directa entre Anbu y Akatsuki
	q, _ = ch.QueueDeclare(
		"localizar_akatsuki", // name
		false,  			  		 // durable
		false,  			  		 // delete when unused
		false,  			  		 // exclusive
		false,  			  		 // no-wait
		nil,    			  		 // arguments
	)

	log.Printf("Declared queue: %s", q.Name)
}

/*
	notificarAkatsuki(data *pb.DataAkatsukiRequest, conn *amqp091.Connection) error
- Envia el Akatsuki recivido al intercambiador y a Hokage a través de RabbitMQ.
*/
func notificarAkatsuki(data *pb.DataAkatsukiRequest, conn *amqp091.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	body, err := proto.Marshal(data)
	if err != nil {
		return err
	}

	err = ch.Publish(
		"EquiposNinjaBrodcast",     	   // exchange
		"", 							   // routing key
		false,  			               // mandatory
		false,  			               // immediate
		amqp091.Publishing{
			ContentType: "application/protobuf",
			Body:        body,
		})
	if err != nil {
		return err
	}

	err = ch.Publish(
		"",     			  		 // exchange
		"notificar_akatsuki_Hokage", // routing key
		false,  			         // mandatory
		false,  			         // immediate
		amqp091.Publishing{
			ContentType: "application/protobuf",
			Body:        body,
		})
	if err != nil {
		return err
	}

	log.Println("Notificación enviada a Hokage y Equipos Ninja")

	return nil
}

/*
	localizarAkatsuki(conn *amqp091.Connection)
- Escucha mensajes en la cola "localizar_akatsuki" de RabbitMQ.
- En cada mensaje recibido, deserializa el contenido y muestra la información del Akatsuki detectado.
- Envía una notificación a Hokage y Equipos Ninja con la información del Akatsuki detectado para que puedan actualizar su información local.
*/
func localizarAkatsuki(conn *amqp091.Connection) {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		"localizar_akatsuki", // queue
		"",     			  // consumer
		true,   			  // auto-ack
		false,  			  // exclusive
		false,  			  // no-local
		false,  			  // no-wait
		nil,    			  // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

    for d := range msgs {
        var data pb.DataAkatsukiRequest

        err := proto.Unmarshal(d.Body, &data)
        if err != nil {
            log.Println("decode error:", err)
            continue
        }

		// Work Here
        log.Printf("Detectado estado Akatsuki >> Nombre: %s, Estado: %s\n", data.Nombre, data.Estado)
		notificarAkatsuki(&data, conn)
    }
}

/*
	main()
- Inicializa la conexión con RabbitMQ, incializa el nodo y la tarea de escuchar mensajes de RabbitMQ.
*/
func main() {
	name := os.Getenv("NAME")
	port := os.Getenv("PORT")
	log.Printf("%s is running on port %s", name, port)

	forever := make(chan bool)

	conn, err := dialRabbitMQ()
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	initialize(conn)

	// Tarea de Recibir mensajes de RabbitMQ desde Akatsuki
	go localizarAkatsuki(conn)

	log.Printf("To exit press CTRL+C")
	<-forever
}