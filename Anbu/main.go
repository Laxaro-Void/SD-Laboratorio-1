package main

import (
	"context"
	"log"
	"time"

	pb "client/proto/homework-system/proto"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("server-container:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("❌ No se pudo conectar al servidor gRPC: %v", err)
	}
	defer conn.Close()
	c := pb.NewHomeworkServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tasks := []string{
		"Informe de Estructuras de Datos",
		"Proyecto de Redes de Computadores",
		"Certamen de Computacion Cientifica",
		"Tarea de Algoritmos",
		"Entrega MVP para Feria",
	}

	for _, task := range tasks {
		resp, err := c.SubmitHomework(ctx, &pb.HomeworkRequest{
			StudentName: "Tralalero Tralala",
			Title:       task,
		})
		if err != nil {
			log.Printf("❌ Error al enviar tarea '%s': %v", task, err)
			continue
		}
		log.Printf("✅ Servidor respondió: %s", resp.Status)
	}
}
