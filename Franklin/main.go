package main

import (
	"log"
	"net"

	pb "server.com/franklin/proto"
	"server.com/franklin/server"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Error escuchando: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterHeistServiceServer(grpcServer, &server.FranklinServer{})

	log.Println("Lester gRPC server escuchando en puerto 50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Error al iniciar gRPC: %v", err)
	}
}
