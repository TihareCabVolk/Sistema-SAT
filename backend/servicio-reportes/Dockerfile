FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/servicio .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/servicio /servicio
EXPOSE 4001
CMD ["/servicio"]
