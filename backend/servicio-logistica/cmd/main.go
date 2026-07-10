package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"servicio-logistica/config"
	"servicio-logistica/consumer"
	"servicio-logistica/handler"
	"servicio-logistica/publisher"
	"servicio-logistica/repository"
	"servicio-logistica/router"
	"servicio-logistica/service"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	gin.SetMode(gin.ReleaseMode)

	db, err := sql.Open("postgres", cfg.DBLogisticaURL)
	if err != nil {
		log.Fatalf("error conectando a DB: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		log.Fatalf("error verificando DB: %v", err)
	}

	if _, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS alertas (
			id VARCHAR(36) PRIMARY KEY,
			id_validacion VARCHAR(36) NOT NULL,
			magnitud DOUBLE PRECISION NOT NULL,
			epicentro_lat DOUBLE PRECISION NOT NULL,
			epicentro_lon DOUBLE PRECISION NOT NULL,
			nivel_alerta VARCHAR(20) NOT NULL,
			costo_emergencia DOUBLE PRECISION NOT NULL,
			zonas_afectadas JSONB DEFAULT '[]',
			estado VARCHAR(20) DEFAULT 'EMITIDA',
			creado_en TIMESTAMP NOT NULL
		);`); err != nil {
		log.Fatalf("error creando tabla alertas: %v", err)
	}

	rmqConn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("error conectando a RabbitMQ: %v", err)
	}
	defer rmqConn.Close()

	repo := repository.NewAlertaRepository(db)
	svc := service.NewLogisticaService(repo)
	pub := publisher.NewRabbitPublisher(rmqConn, cfg.RabbitMQExchange)
	h := handler.NewHandler(svc)

	cons := consumer.NewRabbitConsumer(rmqConn, cfg.RabbitMQExchange, svc, pub)
	go func() {
		log.Println("[main] iniciando consumer RabbitMQ...")
		if err := cons.Start(context.Background()); err != nil {
			log.Printf("[main] consumer terminó: %v", err)
		}
	}()

	r := router.SetupRouter(h)
	go func() {
		addr := fmt.Sprintf(":%s", cfg.ServerPort)
		log.Printf("[main] servidor HTTP en %s", addr)
		if err := r.Run(addr); err != nil {
			log.Fatalf("error en servidor: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[main] apagando...")
}
