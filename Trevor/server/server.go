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

/*
func (s *TrevorServer) RealizarDistraccion(ctx context.Context, req *pb.Solicitud) (*pb.Resultado, error) 
Resumen:
	Realiza la distracción durante la misión.
Simula una serie de turnos basados en la probabilidad de Trevor.
Si en la mitad de los turnos ocurre un evento adverso (10% de probabilidad), la misión falla.
Parámetros:
  - ctx: contexto de la solicitud gRPC.
  - req: estructura Solicitud con la probabilidad de Trevor.
Retorna:
  - Resultado: estructura con el resultado de la distracción (éxito o fallo).
  - error: error en caso de fallo en el proceso.
*/
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

/*
func (s *TrevorServer) PagarParte(ctx context.Context, monto *pb.Monto) (*pb.Confirmacion, error) 
Resumen:
	Recibe el pago de la parte del botín correspondiente a Trevor.
Verifica que el monto sea válido (>0) y responde con una confirmación.
Parámetros:
  - ctx: contexto de la solicitud gRPC.
  - monto: estructura Monto con la cantidad a pagar.
Retorna:
  - Confirmacion: estructura con el resultado del pago y un mensaje de respuesta.
  - error: error en caso de fallo en el proceso.
*/
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
