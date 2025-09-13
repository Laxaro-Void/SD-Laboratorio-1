package server

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	pb "server.com/trevor/proto"
)

type TrevorServer struct {
	pb.UnimplementedHeistTrevorServiceServer
	mu sync.Mutex
}

func (s *TrevorServer) RealizarDistraccion(ctx context.Context, req *pb.Solicitud) (*pb.Resultado, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rand.Seed(time.Now().UnixNano())

	turnos := int(200 - req.ProbTrevor)

	for i := 1; i <= turnos; i++ {
		if turnos/2 == i {
			if rand.Intn(100) <= 10 {
				fmt.Printf("Turno %d: A Trevor le gustaron tanto los terremotos que se quedó prácticamente tieso en el suelo, ha fallado la mision.\n", i)
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

func (s *TrevorServer) PagarParte(ctx context.Context, monto *pb.Monto) (*pb.Confirmacion, error) {

	// Verificar si el monto es válido (>0)
	if monto.Cantidad <= 0 {
		return &pb.Confirmacion{
			Correcto:  false,
			Respuesta: "No recibi el pago correcto >:(",
		}, nil
	}

	msg := fmt.Sprintf("A la p%$@ hora mikey! ... Pero hicimos un buen trabajo.")

	return &pb.Confirmacion{
		Correcto:  true,
		Respuesta: msg,
	}, nil
}
