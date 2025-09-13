package server

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	pb "server.com/franklin/proto"
)

type FranklinServer struct {
	pb.UnimplementedHeistFranklinServiceServer
	mu sync.Mutex
}

func (s *FranklinServer) RealizarDistraccion(ctx context.Context, req *pb.Solicitud) (*pb.Resultado, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rand.Seed(time.Now().UnixNano())

	turnos := int(200 - req.ProbFranklin)

	for i := 1; i <= turnos; i++ {
		if turnos/2 == i {
			if rand.Intn(100) <= 10 {
				fmt.Printf("Turno %d: Franklin ha sido descubierto en las cercanias del lugar debido a que su perro Chop ha ladrado mucho, abortando mision...\n", i)
				return &pb.Resultado{Res: false}, nil
			}
		}
		fmt.Printf("Turno %d: \n", i)
		time.Sleep(50 * time.Millisecond) // 50ms por turno
	}

	resultado := &pb.Resultado{
		Res: true,
	}

	return resultado, nil
}
func (s *FranklinServer) PagarParte(ctx context.Context, monto *pb.Monto) (*pb.Confirmacion, error) {

	// Verificar si el monto es válido (>0)
	if monto.Cantidad <= 0 {
		return &pb.Confirmacion{
			Correcto:  false,
			Respuesta: "No recibi el pago correcto >:(",
		}, nil
	}

	msg := fmt.Sprintf("Perfecto! Cada vez mas cerca de la cima") //mensaje de Trevor "A la %$@r hora mikey! ... Pero hicimos un buen trabajo."

	return &pb.Confirmacion{
		Correcto:  true,
		Respuesta: msg,
	}, nil
}
