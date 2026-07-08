# Multi-stage build: cumple el requisito de "imágenes optimizadas y ligeras (Alpine)".
# Etapa 1 compila; etapa 2 solo contiene el binario (~15 MB final vs ~800 MB con todo Go).
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# CGO_ENABLED=0 -> binario estático, no necesita libc del sistema
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o servicio3 ./cmd

FROM alpine:3.20
RUN adduser -D -u 1000 appuser
USER appuser
WORKDIR /app
COPY --from=builder /app/servicio3 .
EXPOSE 8083
ENTRYPOINT ["./servicio3"]
