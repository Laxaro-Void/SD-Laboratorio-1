package server

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	pb "server.com/lester/proto"
)

type LesterServer struct {
	pb.UnimplementedHeistServiceServer
	mu       sync.Mutex
	rechazos int
}

func (s *LesterServer) SolicitarOferta(ctx context.Context, req *pb.Solicitud) (*pb.Oferta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Peticion de oferta recibida")

	rand.Seed(time.Now().UnixNano())

	// Probabilidad oferta
	if rand.Intn(100) >= 90 {
		return &pb.Oferta{Disponible: false}, nil
	}

	// Espera de 10 segundos
	if s.rechazos >= 3 {
		log.Printf("Se rechazaron 3 ofertas simultaneamente, así que se enojó D:< ")
		time.Sleep(10 * time.Second)
		s.rechazos = 0
	}

	oferta := &pb.Oferta{
		Botin:          int32(rand.Intn(1_000_000) + 100_000),
		ProbFranklin:   int32(rand.Intn(101)),
		ProbTrevor:     int32(rand.Intn(101)),
		RiesgoPolicial: int32(rand.Intn(101)),
		Disponible:     true,
	}
	return oferta, nil
}

func (s *LesterServer) AceptarOferta(ctx context.Context, dec *pb.Decision) (*pb.Respuesta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if dec.Aceptada {
		s.rechazos = 0
		return &pb.Respuesta{Mensaje: "Atraco aceptado por Michael"}, nil
	}
	s.rechazos++
	return &pb.Respuesta{Mensaje: "Oferta rechazada"}, nil
}
