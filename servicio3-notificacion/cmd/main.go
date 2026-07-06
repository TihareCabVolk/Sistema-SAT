package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq" // driver postgres para database/sql

	"github.com/kundev/servicio3-notificacion/internal/broker"
	"github.com/kundev/servicio3-notificacion/internal/config"
	"github.com/kundev/servicio3-notificacion/internal/repository"
	"github.com/kundev/servicio3-notificacion/internal/service"
)

func main() {
	cfg := config.Load()

	// Contexto raíz cancelable por señales del sistema.
	// En Kubernetes, al hacer rollout o matar el pod llega un SIGTERM:
	// cancelamos el consumo, terminamos el mensaje en curso y salimos limpio.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// --- Base de datos (con reintentos, igual que el broker) ---
	db := conectarDB(cfg.DatabaseURL)
	defer db.Close()

	// --- Broker ---
	mq, err := broker.Connect(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("no se pudo conectar a RabbitMQ: %v", err)
	}
	defer mq.Close()

	// --- Wiring (inyección de dependencias manual) ---
	repo := repository.NewAlertaRepository(db)
	svc := service.NewNotificacionService(repo, mq)

	// --- Health check para liveness/readiness probes de Kubernetes ---
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
			if err := db.Ping(); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
		log.Printf("[http] /health en :%s", cfg.HTTPPort)
		if err := http.ListenAndServe(":"+cfg.HTTPPort, nil); err != nil {
			log.Printf("[http] servidor detenido: %v", err)
		}
	}()

	// --- Loop principal de consumo (bloqueante) ---
	if err := mq.Consumir(ctx, svc.ProcesarSismoValidado); err != nil {
		log.Fatalf("consumo terminó con error: %v", err)
	}
	log.Println("servicio 3 detenido correctamente")
}

func conectarDB(url string) *sql.DB {
	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Fatalf("configuración de BD inválida: %v", err)
	}
	for i := 1; i <= 10; i++ {
		if err = db.Ping(); err == nil {
			log.Println("[db] conexión establecida")
			return db
		}
		log.Printf("[db] intento %d/10 fallido: %v — reintentando en 3s", i, err)
		time.Sleep(3 * time.Second)
	}
	log.Fatalf("no se pudo conectar a PostgreSQL: %v", err)
	os.Exit(1)
	return nil
}
