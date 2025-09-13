# Imagen oficial de Go 
FROM golang:1.23.12

# Directorio de trabajo dentro del contenedor
WORKDIR /app

# Copiar archivos de dependencias primero
COPY Trevor/go.mod Trevor/go.sum ./
RUN go mod download

# Copiar el resto del código fuente
COPY Trevor ./

# Compilar el binario
RUN go build -o trevor main.go

# Puerto expuesto (solo para el servidor gRPC)
EXPOSE 50053

# Comando de inicio
CMD ["./trevor"]