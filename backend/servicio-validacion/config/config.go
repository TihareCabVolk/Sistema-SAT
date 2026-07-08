package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort        string
	DBValidacionURL   string
	RabbitMQURL       string
	RabbitMQExchange  string
	ColaSenales       string
	RadioKm           float64
	VentanaSeg        int
	MinSensores       int
}

func Load() *Config {
	return &Config{
		ServerPort:       getEnv("SERVER_PORT", "4002"),
		DBValidacionURL:  getEnv("DB_VALIDACION_URL", "postgres://postgres:postgres@localhost:5432/validacion?sslmode=disable"),
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672"),
		RabbitMQExchange: getEnv("RABBITMQ_EXCHANGE", "sat.events"),
		ColaSenales:      getEnv("COLA_SENALES", "cola_señales_recibidas"),
		RadioKm:          getEnvFloat("RADIO_KM", 100),
		VentanaSeg:       getEnvInt("VENTANA_SEG", 60),
		MinSensores:      getEnvInt("MIN_SENSORES", 3),
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			return n
		}
	}
	return defaultValue
}
