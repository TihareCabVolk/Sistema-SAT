package rabbitmq

import (
	"context"
	"fmt"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var Canal *amqp.Channel
var Conn *amqp.Connection
var Exchange string

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func InitRabbit() error {
	var err error

	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	Exchange = getEnv("RABBITMQ_EXCHANGE", "sat.events")

	for i := 1; i <= 5; i++ {
		Conn, err = amqp.Dial(rabbitURL)
		if err == nil {
			fmt.Println("Conexión exitosa a RabbitMQ")
			break
		}
		fmt.Printf("Esperando a RabbitMQ... (Intento %d/5)\n", i)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		return err
	}

	Canal, err = Conn.Channel()
	if err != nil {
		return err
	}

	err = Canal.ExchangeDeclare(
		Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}

func Close() {
	if Canal != nil {
		Canal.Close()
	}
	if Conn != nil {
		Conn.Close()
	}
}

func PublicarEvento(cola string, payload []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := Canal.PublishWithContext(ctx,
		Exchange,
		cola,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         payload,
		})
	return err
}
