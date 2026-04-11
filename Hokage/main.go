package main

import (
	// "context"
	// "fmt"
	"context"
	"log"
	"net"

	// "net"
	"os"
	// "time"

	pb "Hokage/proto/message"

	"google.golang.org/grpc"
	//"github.com/rabbitmq/amqp091-go"
)

type Akatsukis struct {
	Id            int
	Nombre        string
	Ataque        int
	Vida          int
	Estado        string
	NinjaAsociado string //Mantendrá el registro de quien es el equipo ninja que se esta encargando del enemigo
	Recompensa    int    //recompensa asociada al enemigo
}

type HokageServer struct {
	pb.UnimplementedMessengerServer
	AkatsukisList []Akatsukis
}

func (h *HokageServer) ObtenerListaAkatsukis(ctx context.Context, req *pb.Empty) (*pb.listaAkatsukis, error) {
	var enemigos []*pb.AkatsukiInfo
	for _, a := range h.AkatsukisList {
		enemigos = append(enemigos, &pb.AkatsukiInfo{
			Nombre: a.Nombre,
			Ataque: a.Ataque,
			Vida:   a.Vida,
			Estado: a.Estado,
			//NinjaAsociado: a.NinjaAsociado,
			Recompensa: a.Recompensa,
		})
	}
	return &pb.listaAkatsukis{Enemigos: enemigos}, nil
}

func main() {
	// Creacion de instancia de servidor
	name := os.Getenv("NAME")
	port := os.Getenv("PORT")
	log.Printf("%s is running on port %s", name, port)

	hokage := &HokageServer{
		AkatsukisList: []Akatsukis{},
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("No se pudo abrir el puerto %s: %v", port, err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterHokageServer(grpcServer, hokage)
	grpcServer.Serve(lis)

	log.Printf("%s corriendo en puerto %s", name, port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Error al servir gRPC: %v", err)
	}

	// logica tareas, instanciacion de funciones
}
