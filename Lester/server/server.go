package server

import (
	"bufio"
	"context"
	"crypto/internal/fips140/edwards25519/field"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	pb "server.com/lester/proto"
)

type LesterServer struct {
	pb.UnimplementedHeistServiceServer
	mu       sync.Mutex
	rechazos int
	index	 int
	file     *os.File

}

/*
function NewLesterServer() *LesterServer
Resumen:
	Crea una nueva instancia del servidor de Lester.
Abre el archivo "ofertas_grandes.csv" para leer las ofertas de atraco.
Inicializa el contador de rechazos y el índice de lectura.
Retorna:
  - *LesterServer: puntero a la nueva instancia del servidor de Lester.
*/
func NewLesterServer() *LesterServer {
	file, err := os.Open("ofertas_grandes.csv")
	if err != nil {
		log.Fatalf("Error al abrir el archivo: %v", err)
	}

	return &LesterServer{
		rechazos: 0,
		index:    0,
		file:     file,
	}
}

/*
func (s *LesterServer) SolicitarOferta(ctx context.Context, req *pb.Solicitud) (*pb.Oferta, error)
Resumen:
	Genera y envía una oferta de atraco a Michael.
La oferta se basa en un archivo CSV con posibles ofertas.
Si se han rechazado 3 ofertas consecutivas, espera 10 segundos antes de generar una nueva.
Existe una probabilidad del 10% de que no haya oferta disponible.
Parámetros:
  - ctx: contexto de la solicitud gRPC.
  - req: estructura Solicitud (actualmente no utilizada).
Retorna:
  - Oferta: estructura con los detalles de la oferta (botín, probabilidades, riesgo, disponibilidad).
  - error: error en caso de fallo en el proceso.
*/
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

	// Leer oferta del archivo
	scanner := bufio.NewScanner(s.file)
	if s.index == 0 {
		scanner.Scan() // Saltar la primera línea (encabezados)
	}
	line := scanner.Text()
	s.index++

	if strings.TrimSpace(line) == "" {
		s.file.Seek(0, 0)
		scanner = bufio.NewScanner(s.file)
		scanner.Scan() // Saltar la primera línea (encabezados)
		line = scanner.Text()
	}

	fields := strings.Split(line, ",")

	var botin, probFranklin, probTrevor, riesgoPolicial int32
	if len(fields) > 0 && strings.TrimSpace(fields[0]) != "" {
		fmt.Sscanf(fields[0], "%d", &botin)
	}
	if len(fields) > 1 && strings.TrimSpace(fields[1]) != "" {
		fmt.Sscanf(fields[1], "%d", &probFranklin)
	}
	if len(fields) > 2 && strings.TrimSpace(fields[2]) != "" {
		fmt.Sscanf(fields[2], "%d", &probTrevor)
	}
	if len(fields) > 3 && strings.TrimSpace(fields[3]) != "" {
		fmt.Sscanf(fields[3], "%d", &riesgoPolicial)
	}

	oferta := &pb.Oferta{
		Botin:          botin,
		ProbFranklin:   probFranklin,
		ProbTrevor:     probTrevor,
		RiesgoPolicial: riesgoPolicial,
		Disponible:     true,
	}
	return oferta, nil
}

/*
func (s *LesterServer) AceptarOferta(ctx context.Context, dec *pb.Decision) (*pb.Respuesta, error)
Resumen:
	Recibe la decisión de Michael sobre la oferta.
Si la oferta es aceptada, resetea el contador de rechazos y confirma la aceptación.
Si es rechazada, incrementa el contador de rechazos y confirma el rechazo.
Parámetros:
  - ctx: contexto de la solicitud gRPC.
  - dec: estructura Decision con la decisión de Michael (aceptada o no).
Retorna:
  - Respuesta: estructura con un mensaje confirmando la decisión.
  - error: error en caso de fallo en el proceso.
*/
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

/*
func (s *LesterServer) PagarParte(ctx context.Context, monto *pb.Monto) (*pb.Confirmacion, error)
Resumen:
	Recibe el pago de la parte del botín correspondiente a Lester.
Verifica que el monto sea válido (>0) y responde con una confirmación.
Parámetros:
  - ctx: contexto de la solicitud gRPC.
  - monto: estructura Monto con la cantidad a pagar.
Retorna:
  - Confirmacion: estructura con el resultado del pago y un mensaje de respuesta.
  - error: error en caso de fallo en el proceso.
*/
func (s *LesterServer) PagarParte(ctx context.Context, monto *pb.Monto) (*pb.Confirmacion, error) {

	// Verificar si el monto es válido (>0)
	if monto.Cantidad <= 0 {
		return &pb.Confirmacion{
			Correcto:  false,
			Respuesta: "No recibi el pago correcto >:(",
		}, nil
	}

	msg := fmt.Sprintf("Un placer hacer negocios. AHORA QUIEN DIJO TEQUILA!!")
	return &pb.Confirmacion{
		Correcto:  true,
		Respuesta: msg,
	}, nil
}
