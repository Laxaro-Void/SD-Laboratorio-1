package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	pb "Hokage/proto/message"

	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedMessengerServer
}

func (s *server) SendMessage(ctx context.Context, req *pb.MessageRequest) (*pb.MessageReply, error) {
	log.Printf("Received via gRPC: %s -> %s", req.From, req.Body)
	return &pb.MessageReply{Status: "OK"}, nil
}

func listenRabbit(name string) {
	conn, err := amqp091.Dial("amqp://guest:guest@" + os.Getenv("RABBITMQ-IP") + "/")
	if err != nil {
		log.Fatal(err)
	}

	ch, _ := conn.Channel()

	q, _ := ch.QueueDeclare(name, false, false, false, false, nil)

	msgs, _ := ch.Consume(q.Name, "", true, false, false, false, nil)

	go func() {
		for msg := range msgs {
			log.Printf("[%s] Received via RabbitMQ: %s", name, string(msg.Body))
		}
	}()

	select {}
}

func sendRabbit(queue, msg string) {
	conn, _ := amqp091.Dial("amqp://guest:guest@" + os.Getenv("RABBITMQ-IP") + "/")
	ch, _ := conn.Channel()

	ch.QueueDeclare(queue, false, false, false, false, nil)

	ch.Publish("", queue, false, false,
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		})
}

func sendGRPC(addr string, msg string) {
	conn, _ := grpc.Dial(addr, grpc.WithInsecure())
	defer conn.Close()

	client := pb.NewMessengerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client.SendMessage(ctx, &pb.MessageRequest{
		From: "sender",
		Body: msg,
	})
}

func sender() {
	nodes := []string{
		os.Getenv("Akatsuki-IP"),
		os.Getenv("Anbu-IP"),
		os.Getenv("EquiposNinja-IP"),
	}

	for _, n := range nodes {
		sendRabbit(n, "Hello via RabbitMQ -> "+n)
		sendGRPC(n, "Hello via gRPC -> "+n)
	}

	log.Println("Messages sent")
}

func main() {
	name := os.Getenv("NAME")
	port := os.Getenv("PORT")

	go sender()
	go listenRabbit(name)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterMessengerServer(s, &server{})

	fmt.Println("Node running on port", port)
	s.Serve(lis)
}