package main

import (
	"context"
	"log"
	"time"

	pb "server.com/lester/proto"

	"google.golang.org/grpc"
)

func main() {
	// Conectar al servidor gRPC de Lester
	conn, err := grpc.Dial("lester:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("No se pudo conectar a Lester: %v", err)
	}
	defer conn.Close()

	client := pb.NewHeistServiceClient(conn)

	valor := BotinTotal / 4
	resto := BotinTotal % 4

	valorlester := valor + resto

	// a cada uno se le envia el valor , pero a lester se el envia valorlester
	// no se como el Cris implemento la conexion para cada servidor, por eso solo lo hice con lester aqui

	// enviar un monto
	monto := &pb.Monto{
		cantidad: valorlester,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	Confirmar, err := client.PagarParte(ctx, monto)
	if err != nil {
		log.Fatalf("Error al enviar parte: %v", err)
	}

	log.Printf("Pago enviado! : Pago correcto =%v, Mensaje agregado =%s",
		Confirmar.correcto, Confirmar.respuesta)
}
