# Imagen oficial de Go 
FROM golang:1.23.12

# Directorio de trabajo dentro del contenedor
WORKDIR /app

# Copiar archivos de dependencias primero
COPY Lester/go.mod Lester/go.sum ./
RUN go mod download

# Copiar el resto del código fuente
COPY Lester ./

# Compilar el binario
RUN go build -o lester main.go

# Puerto expuesto (solo para el servidor gRPC)
EXPOSE 50051

# Comando de inicio
CMD ["./lester"]