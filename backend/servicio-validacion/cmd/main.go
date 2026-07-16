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
		db, innerErr = repository.InitDB(cfg.DBValidacionURL)
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

	repo := repository.NewSenalRepository(db.DB)
	svc := service.NewValidacionService(repo, cfg.RadioKm, cfg.VentanaSeg, cfg.MinSensores)
	pub := publisher.NewRabbitPublisher(rmqConn, cfg.RabbitMQExchange)

	cons := consumer.NewRabbitConsumer(rmqConn, cfg.ColaSenales, cfg.RabbitMQExchange, svc, pub)
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

	r := router.SetupRouter()
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
	rmqConn.Close()
	repo.Close()
}
