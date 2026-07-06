package config

import "os"

type Config struct {
	ServerPort       string
	DBLogisticaURL   string
	RabbitMQURL      string
	RabbitMQExchange string
}

func Load() *Config {
	return &Config{
		ServerPort:       getEnv("SERVER_PORT", "4003"),
		DBLogisticaURL:   getEnv("DB_LOGISTICA_URL", "postgres://postgres:postgres@localhost:5432/logistica?sslmode=disable"),
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672"),
		RabbitMQExchange: getEnv("RABBITMQ_EXCHANGE", "sat.events"),
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
