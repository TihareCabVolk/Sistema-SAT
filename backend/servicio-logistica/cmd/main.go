package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
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

func retryWithBackoff(ctx context.Context, maxRetries int, base time.Duration, fn func() error) error {
	var err error
	for i := range maxRetries {
		if err = fn(); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(base * (1 << i)):
		}
	}
	return err
}

func main() {
	cfg := config.Load()

	gin.SetMode(gin.ReleaseMode)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var db *repository.SafeDB
	backoffCtx, backoffCancel := context.WithTimeout(ctx, 30*time.Second)
	defer backoffCancel()

	if err := retryWithBackoff(backoffCtx, 5, time.Second, func() error {
		var innerErr error
		db, innerErr = repository.InitDB(cfg.DBLogisticaURL)
		return innerErr
	}); err != nil {
		log.Fatalf("error conectando a DB: %v", err)
	}

	var rmqConn *amqp.Connection
	if err := retryWithBackoff(backoffCtx, 5, time.Second, func() error {
		var innerErr error
		rmqConn, innerErr = amqp.Dial(cfg.RabbitMQURL)
		return innerErr
	}); err != nil {
		log.Fatalf("error conectando a RabbitMQ: %v", err)
	}

	repo := repository.NewAlertaRepository(db.DB)
	svc := service.NewLogisticaService(repo)
	pub := publisher.NewRabbitPublisher(rmqConn, cfg.RabbitMQExchange)
	h := handler.NewHandler(svc)

	cons := consumer.NewRabbitConsumer(rmqConn, cfg.RabbitMQExchange, svc, pub)
	consCtx, consCancel := context.WithCancel(ctx)
	defer consCancel()

	go func() {
		for {
			log.Println("[main] iniciando consumer RabbitMQ...")
			if err := cons.Start(consCtx); err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				log.Printf("[main] consumer terminó: %v. Reintentando en 3s...", err)
				time.Sleep(3 * time.Second)
				continue
			}
			break
		}
	}()

	r := router.SetupRouter(h)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
		Handler: r,
	}

	go func() {
		log.Printf("[main] servidor HTTP en :%s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("error en servidor: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[main] apagando...")

	consCancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[main] error apagando servidor: %v", err)
	}

	if err := db.Close(); err != nil {
		log.Printf("[main] error cerrando DB: %v", err)
	}
	rmqConn.Close()
}
