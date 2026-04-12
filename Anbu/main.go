package main

import (
	// "context"
	"fmt"
	"log"
	// "net"
	"os"
	"time"

	pb "Anbu/proto/DataAkatsuki"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

type Akatsukis struct {
	Id int
	Name string
	Ataque int
	Vida int
	Estado string
	isComunicated bool
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

/**
* Crea el canal "notificar_akatsuki" en RabbitMQ para enviar notificaciones a Akatsuki
**/
func initialize() {
	conn, err := dialRabbitMQ()
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()
	
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
		return
	}
	defer ch.Close()

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

	q, _ := ch.QueueDeclare(
		"notificar_akatsuki_Hokage", // name
		false,  			  		 // durable
		false,  			  		 // delete when unused
		false,  			  		 // exclusive
		false,  			  		 // no-wait
		nil,    			  		 // arguments
	)

	log.Printf("Declared queue: %s", q.Name)

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

func notificarAkatsuki(data *pb.DataAkatsukiRequest) error {
	conn, err := dialRabbitMQ()
	if err != nil {
		return err
	}
	defer conn.Close()
	
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

	log.Printf("Sent: %+v\n", data.Id)
	return nil
}

/**
* Recibe Akattsuki desde RabbitMQ desde el canal "localizar_akatsuki"
* Origen: Akatsuki 
**/
func localizarAkatsuki() {
	conn, err := dialRabbitMQ()
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

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
        log.Printf("Received: %+v\n", data.Id)
		notificarAkatsuki(&data)
    }
}


func main() {
	name := os.Getenv("NAME")
	port := os.Getenv("PORT")
	log.Printf("%s is running on port %s", name, port)

	forever := make(chan bool)

	initialize()

	// Tarea de Recivir mensajes de RabbitMQ desde Akatsuki
	go localizarAkatsuki()

	log.Printf("Waiting for messages. To exit press CTRL+C")
	<-forever
}