package config

import (
	"log"
	"os"
)

type Config struct {
	RabbitMQURL string
	DatabaseURL string
	HTTPPort    string
}

func Load() Config {
	cfg := Config{
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://s3:s3pass@localhost:5433/notificaciones?sslmode=disable"),
		HTTPPort:    getEnv("HTTP_PORT", "8083"),
	}
	log.Printf("[config] broker=%s puerto_http=%s", cfg.RabbitMQURL, cfg.HTTPPort)
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
