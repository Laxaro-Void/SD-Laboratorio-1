package main

import (
	"log"
	"os"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	// Conexion Basica
	url := os.Getenv("RABBITMQ_URL")
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf(" Error al conectar : %v", err)
	}
	defer conn.Close()

	// Apertura de canal
	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf(" Error al abrir el canal: %v", err)
	}
	defer channel.Close()

	// Declaracion de la cola de mensajes
	queue, err := channel.QueueDeclare(
		"tareas", // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // args
	)
	if err != nil {
		log.Fatalf(" Error al declarar la cola de mensajes: %v", err)
	}
	defer conn.Close()

	// Consumir mensajes de la cola
	msgs, err := channel.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		log.Fatalf(" Error al consumir la cola de mensajes: %v", err)
	}
	defer conn.Close()

	log.Println("🟡 Esperando tareas...")
	for msg := range msgs {
		log.Printf("✅ Procesando: %s", msg.Body)
	}
}