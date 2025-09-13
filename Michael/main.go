package main

import (
	"context"
	"log"
	"time"
	"os"
	"fmt"

	pbFranklin "server.com/franklin/proto"
	pb "server.com/lester/proto"
	pbTrevor "server.com/trevor/proto"

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

	// ------------------------------ Parte 1 ------------------------------

	// Solicitar una oferta
	solicitud := &pb.Solicitud{}

	aceptado := false
	var oferta *pb.Oferta

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

		if oferta.Botin == 0 || oferta.ProbFranklin == 0 || oferta.ProbTrevor == 0 || oferta.RiesgoPolicial == 0 {
			log.Println("Algún campo de la oferta es nulo o cero, Rechazado...")
			continue
		}

		log.Printf("Oferta recibida: Botín=%d, Franklin=%d, Trevor=%d, RiesgoPolicial=%d, Disponible=%v",
			oferta.Botin, oferta.ProbFranklin, oferta.ProbTrevor, oferta.RiesgoPolicial, oferta.Disponible)

		// Decidir aceptar o rechazar
		decision := &pb.Decision{
			Aceptada: !(oferta.Botin == 0 || oferta.ProbFranklin == 0 || oferta.ProbTrevor == 0 || oferta.RiesgoPolicial == 0) && oferta.Disponible && oferta.RiesgoPolicial < 80 && (oferta.ProbFranklin > 50 || oferta.ProbTrevor > 50),
		}
		aceptado = decision.Aceptada

		respuesta, err := client.AceptarOferta(ctx, decision)
		if err != nil {
			log.Fatalf("Error al aceptar oferta: %v", err)
		}
		log.Printf("Respuesta de Lester: %s", respuesta.Mensaje)
		time.Sleep(100 * time.Millisecond)
	}

	// ------------------------------ Parte 2 ------------------------------

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

	} else {
		solicitud := &pbTrevor.Solicitud{
			ProbTrevor: oferta.ProbTrevor,
		}
		resultado, err := clientTrevor.RealizarDistraccion(ctxT, solicitud)

		if err != nil {
			log.Fatalf("Error al realizar distracción: %v", err)
		}
		log.Printf("Distracción de Franklin resultado: %v", resultado.Res)
	}

	// ------------------------------ Parte 3 ------------------------------

	// YOUR OWN CODE

	// ------------------------------ Parte 4 ------------------------------

	valor := BotinTotal / 4
	resto := BotinTotal % 4

	valorlester := valor + resto

	// a cada uno se le envia el valor , pero a lester se el envia valorlester
	// Crear mensaje con el monto para lester
	monto := &pb.Monto{
		Cantidad: valorlester,
	}

	ConfirmarL, err := client.PagarParte(ctx, monto)
	if err != nil {
		log.Fatalf("Error al enviar la parte de Lester: %v", err)
	}

	log.Printf("Pago enviado! : Pago correcto? %v, Nuevo mensaje de Lester Crest: %s",
		ConfirmarL.correcto, ConfirmarL.respuesta)

	// Actualizar el monto para el Franklin y Trevor
	monto := &pb.Monto{
		Cantidad: valor,
	}

	// Crear mensaje con el monto para Franklin
	ConfirmarF, err := clientFranklin.PagarParte(ctx, monto)
	if err != nil {
		log.Fatalf("Error al enviar la parte de Franklin: %v", err)
	}

	log.Printf("Pago enviado! : Pago correcto? %v, Nuevo mensaje de Franklin Clinton: %s",
		ConfirmarF.correcto, ConfirmarF.respuesta)

	// Crear mensaje con el monto para Trevor
	ConfirmarT, err := clientTrevor.PagarParte(ctx, monto)
	if err != nil {
		log.Fatalf("Error al enviar la parte de Trevor: %v", err)
	}

	log.Printf("Pago enviado! : Pago correcto? %v, Nuevo mensaje de Trevor Philips: %s",
		ConfirmarT.correcto, ConfirmarT.respuesta)
}
