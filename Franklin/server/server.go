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
			if rand.Intn(100) >= 10 {
				return &pb.Resultado{res: false}, nil
			}
		}
		fmt.Printf("Turno %d: \n", i)
		time.Sleep(100 * time.Millisecond) // 100ms por turno
	}

	return resultado, nil
}
