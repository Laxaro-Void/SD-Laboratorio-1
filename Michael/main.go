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

	// Solicitar una oferta
	solicitud := &pb.Solicitud{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	aceptado := false

	for !aceptado {

		oferta, err := client.SolicitarOferta(ctx, solicitud)
		if err != nil {
			log.Fatalf("Error al solicitar oferta: %v", err)
		}

		log.Printf("Oferta recibida: Botín=%d, Franklin=%d, Trevor=%d, RiesgoPolicial=%d, Disponible=%v",
			oferta.Botin, oferta.ProbFranklin, oferta.ProbTrevor, oferta.RiesgoPolicial, oferta.Disponible)

		// Decidir aceptar o rechazar
		decision := &pb.Decision{
			Aceptada: oferta.Disponible && oferta.RiesgoPolicial < 80 && (oferta.ProbFranklin > 50 || oferta.ProbTrevor > 50),
		}
		aceptado = decision.Aceptada

		respuesta, err := client.AceptarOferta(ctx, decision)
		if err != nil {
			log.Fatalf("Error al aceptar oferta: %v", err)
		}
		log.Printf("Respuesta de Lester: %s", respuesta.Mensaje)
	}

}
