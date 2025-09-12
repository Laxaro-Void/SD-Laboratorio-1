# Dockerfile
FROM golang:1.25 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Build with static linking
RUN CGO_ENABLED=0 go build -o app

FROM scratch
COPY --from=builder /app/app /

CMD ["/app"]