FROM golang:1.23

# Directorio de trabajo
WORKDIR /app/Michael

# Copiar go.mod y go.sum del cliente
COPY Michael/go.mod Michael/go.sum ./

# Copiar el módulo Lester
COPY Lester ./../Lester

# Descargar dependencias
RUN go mod download

# Copiar todo el código del cliente
COPY Michael .

# Compilar
RUN go build -o michael main.go

# Comando por defecto
CMD ["./michael"]
