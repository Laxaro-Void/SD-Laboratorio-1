package main

import (
	"context"
	"log"
	"time"
	"os"
	"fmt"

	pb "server.com/lester/proto"

	"google.golang.org/grpc"
)

type Reporte struct {
	Mision         string
	Botin          int32
	BotinExtra     int32
	BotinTotal     int32
	BotinFranklin  int32
	BotinTrevor    int32
	BotinLester    int32
	RestoLester    int32
	MensajeFranklin string
	MensajeTrevor   string
	MensajeLester   string
	Estado         bool
}

func generar_reporte_final(reporte Reporte)  {
	// Ver si es crear un único reporte, o es un reporte por cada mision
	file, err := os.Create("Reporte.txt")
	if err != nil {
		log.Fatalf("No se pudo crear el archivo: %v", err)
	}
	defer file.Close()

	file.WriteString("============================================================\n")
	file.WriteString("==               REPORTE FINAL DE LA MISION               ==\n")
	file.WriteString("============================================================\n")
	file.WriteString("\n")
	file.WriteString("Mision: " + reporte.Mision + "\n")
	file.WriteString("Resultado Global: " + func() string {
		if reporte.Estado {
			return "MISION COMPLETADA CON EXITO!"
		}
		return "MISION FALLIDA"
	}() + "\n")
	file.WriteString("\n")
	file.WriteString("            --- REPARTO DEL BOTIN ---            \n")
	if reporte.Estado {
		file.WriteString("Botin Base: $" +  fmt.Sprintf("%d", reporte.Botin) + "\n")
		file.WriteString("Botin Extra: $" +  fmt.Sprintf("%d", reporte.BotinExtra) + "\n")
		file.WriteString("Botin Total: $" +  fmt.Sprintf("%d", reporte.BotinTotal) + "\n")
		file.WriteString("---------------------------------------------------------\n")
		file.WriteString("Pago a Franklin: $" + fmt.Sprintf("%d", reporte.BotinFranklin) + "\n")
		file.WriteString("Respuesta de Franklin: \"" + reporte.MensajeFranklin + "\"\n")
		file.WriteString("\n")
		file.WriteString("Pago a Trevor: $" + fmt.Sprintf("%d", reporte.BotinTrevor) + "\n")
		file.WriteString("Respuesta de Trevor: \"" + reporte.MensajeTrevor + "\"\n")
		file.WriteString("\n")
		file.WriteString("Pago a Lester: $" + fmt.Sprintf("%d", reporte.BotinLester) + " (reparto) + $" + fmt.Sprintf("%d", reporte.RestoLester) + " (resto)\n")
		file.WriteString("Respuesta de Lester: \"" + reporte.MensajeLester + "\"\n")
		file.WriteString("\n")
		file.WriteString("---------------------------------------------------------\n")
		file.WriteString("Saldo Final de la Operación: $" + fmt.Sprintf("%d", reporte.BotinTotal) + "\n")
		file.WriteString("============================================================\n")
	} else {
		file.WriteString("Botin Base: $" +  fmt.Sprintf("%d", reporte.Botin) + "\n")
		file.WriteString("Botin Extra: $" +  fmt.Sprintf("%d", reporte.BotinExtra) + "\n")
		file.WriteString("Botin Perdido: $" + fmt.Sprintf("%d", reporte.BotinTotal) + "\n")

		// Falta reporte de fallo, cual fase, por que fallo, y quienes parciparon
	}

	err = file.Close()
	if err != nil {
		log.Fatalf("Error al cerrar el archivo: %v", err)
	}
}

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
