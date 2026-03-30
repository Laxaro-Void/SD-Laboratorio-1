package main

import (
	// "context"
	// "fmt"
	"log"
	// "net"
	"os"
	// "time"


	//pb "Akatsuki/proto/message"

	//"github.com/rabbitmq/amqp091-go"
	//"google.golang.org/grpc"
)

func main() {
	name := os.Getenv("NAME")
	port := os.Getenv("PORT")

	log.Printf("%s is running on port %s", name, port)
}