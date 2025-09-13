package main

import (
	"log"
	"net"

	pb "server.com/trevor/proto"
	"server.com/trevor/server"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("Error escuchando: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterHeistTrevorServiceServer(grpcServer, &server.TrevorServer{})

	log.Println("Trevor  gRPC server escuchando en puerto 50053")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Error al iniciar gRPC: %v", err)
	}
}
