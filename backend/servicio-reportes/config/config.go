package config

import "os"

type Config struct {
	ServerPort      string
	DBReportesURL   string
	RabbitMQURL     string
	RabbitMQExchange string
}

func Load() *Config {
	return &Config{
		ServerPort:       getEnv("SERVER_PORT", "4001"),
		DBReportesURL:    getEnv("DB_REPORTES_URL", "postgres://reportes:password@localhost:5432/reportes?sslmode=disable"),
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RabbitMQExchange: getEnv("RABBITMQ_EXCHANGE", "sat.events"),
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}