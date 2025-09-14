package main

import (
	"context"
	"log"
	"sync"
	"time"

	pbFranklin "server.com/franklin/proto"
	pb "server.com/lester/proto"
	pbTrevor "server.com/trevor/proto"

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

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Conectar al servidor gRPC de Franklin
	conn, err = grpc.Dial("franklin:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("No se pudo conectar a Franklin: %v", err)
	}
	defer conn.Close()

	clientFranklin := pbFranklin.NewHeistFranklinServiceClient(conn)

	ctxF, cancelF := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelF()

	// Conectar al servidor gRPC de Trevor
	conn, err = grpc.Dial("trevor:50053", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("No se pudo conectar a Trevor: %v", err)
	}
	defer conn.Close()

	clientTrevor := pbTrevor.NewHeistTrevorServiceClient(conn)

	ctxT, cancelT := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelT()

	// Se crea variable donde almacenar la solicitud
	solicitud := &pb.Solicitud{}

	aceptado := false
	var oferta *pb.Oferta

	// Se piden trabajos a Lester hasta que tenga una oferta lo suficientemente buena
	for !aceptado {

		var err error
		oferta, err = client.SolicitarOferta(ctx, solicitud)
		if err != nil {
			log.Fatalf("Error al solicitar oferta: %v", err)
		}
		if oferta == nil {
			log.Println("Lester devolvió una oferta nula, reintentando...")
			continue
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
		time.Sleep(100 * time.Millisecond)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	res := make(chan bool, 1)

	log.Printf("Franklin: %d", oferta.ProbFranklin)
	if oferta.ProbFranklin > oferta.ProbTrevor {
		solicitud := &pbFranklin.Solicitud{
			ProbFranklin: oferta.ProbFranklin,
		}
		resultado, err := clientFranklin.RealizarDistraccion(ctxF, solicitud)

		if err != nil {
			log.Fatalf("Error al realizar distracción: %v", err)
		}
		log.Printf("Distracción de Franklin resultado: %v", resultado.Res)

		go func() {
			defer wg.Done()

		}()

	} else {
		solicitud := &pbTrevor.Solicitud{
			ProbTrevor: oferta.ProbTrevor,
		}
		resultado, err := clientTrevor.RealizarDistraccion(ctxT, solicitud)

		if err != nil {
			log.Fatalf("Error al realizar distracción: %v", err)
		}
		log.Printf("Distracción de Trevor resultado: %v", resultado.Res)

		go func() {
			defer wg.Done()
			resultado, err := clientFranklin.RealizarAtraco(ctxF, &pbFranklin.SolicitudAtraco{ProbFranklin: oferta.ProbFranklin, Botin: oferta.Botin})
			if err != nil {
				log.Fatalf("Error al realizar atraco: %v", err)
			}
			log.Printf("Atraco de Franklin resultado: %v", resultado.Res)
			res <- resultado.Res
		}()
	}

	go func() {
		defer wg.Done()
		_, err := client.EstudiarGolpe(ctx, &pb.Riesgo{Risk: oferta.RiesgoPolicial})
		if err != nil {
			log.Fatalf("Error al llamar a Lester: %v", err)
		}
	}()

	wg.Wait()
}
