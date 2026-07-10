package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"servicio-validacion/config"
	"servicio-validacion/consumer"
	"servicio-validacion/publisher"
	"servicio-validacion/repository"
	"servicio-validacion/router"
	"servicio-validacion/service"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	gin.SetMode(gin.ReleaseMode)

	db, err := repository.InitDB(cfg.DBValidacionURL)
	if err != nil {
		log.Fatalf("error conectando a DB: %v", err)
	}
	defer db.Close()

	rmqConn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("error conectando a RabbitMQ: %v", err)
	}
	defer rmqConn.Close()

	repo := repository.NewSenalRepository(db)
	svc := service.NewValidacionService(repo, cfg.RadioKm, cfg.VentanaSeg, cfg.MinSensores)
	pub := publisher.NewRabbitPublisher(rmqConn, cfg.RabbitMQExchange)

	cons := consumer.NewRabbitConsumer(rmqConn, cfg.ColaSenales, cfg.RabbitMQExchange, svc, pub)
	go func() {
		for {
		log.Println("[main] iniciando consumer RabbitMQ...")
		if err := cons.Start(context.Background()); err != nil {
			log.Printf("[main] consumer terminó: %v. Reintentando en 3s...", err)
			time.Sleep(3 * time.Second)
			continue
		}
		break
	}()

	r := router.SetupRouter()
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
