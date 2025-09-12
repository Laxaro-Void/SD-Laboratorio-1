# Dockerfile
FROM golang:1.23 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o app

FROM debian:bullseye-slim
WORKDIR /root/
COPY --from=builder /app/app .

CMD ["./app"]
