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

	"servicio-reportes/config"
	"servicio-reportes/rabbitmq"
	"servicio-reportes/repository"
	"servicio-reportes/router"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	backoffCtx, backoffCancel := context.WithTimeout(ctx, 30*time.Second)
	defer backoffCancel()

	if err := retryWithBackoff(backoffCtx, 5, time.Second, func() error {
		return repository.InitDB()
	}); err != nil {
		log.Fatalf("error conectando a DB: %v", err)
	}
	defer repository.Client.Close()

	if err := retryWithBackoff(backoffCtx, 5, time.Second, func() error {
		return rabbitmq.InitRabbit()
	}); err != nil {
		log.Fatalf("error conectando a RabbitMQ: %v", err)
	}

	router.SetupRouter()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
		Handler: nil,
	}

	consumidorCtx, consumidorCancel := context.WithCancel(ctx)
	defer consumidorCancel()

	go rabbitmq.StartConsumer(consumidorCtx)

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

	consumidorCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[main] error apagando servidor: %v", err)
	}

	if err := repository.Client.Close(); err != nil {
		log.Printf("[main] error cerrando DB: %v", err)
	}
	rabbitmq.Close()
}
