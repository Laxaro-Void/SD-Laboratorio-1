# Imagen oficial de Go 
FROM golang:1.23.12

# Directorio de trabajo dentro del contenedor
WORKDIR /app

# Copiar archivos de dependencias primero
COPY Franklin/go.mod Franklin/go.sum ./
RUN go mod download

# Copiar el resto del código fuente
COPY Franklin ./

# Compilar el binario
RUN go build -o franklin main.go

# Puerto expuesto (solo para el servidor gRPC)
EXPOSE 50052

# Comando de inicio
CMD ["./franklin"]