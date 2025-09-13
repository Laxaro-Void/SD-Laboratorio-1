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
