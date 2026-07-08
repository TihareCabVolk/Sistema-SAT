package rabbitmq

import (
	"context"
	"fmt"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// via para enviar los mensajes
var Canal *amqp.Channel
var Exchange string

func InitRabbit() {
	var conn *amqp.Connection
	var err error

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	Exchange = os.Getenv("RABBITMQ_EXCHANGE")
	if Exchange == "" {
		Exchange = "sat.events"
	}

	// Intentos
	for i := 1; i <= 5; i++ {
		conn, err = amqp.Dial(rabbitURL)
		if err == nil {
			fmt.Println("Conexión exitosa a RabbitMQ")
			break
		}
		fmt.Printf("Esperando a RabbitMQ... (Intento %d/5)\n", i)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		fmt.Println("No se pudo conectar a RabbitMQ:", err)
		os.Exit(1)
	}

	// Abrir canal de comunicación
	Canal, err = conn.Channel()
	if err != nil {
		fmt.Println("Error abriendo canal de RabbitMQ:", err)
		os.Exit(1)
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
		fmt.Println("Error declarando el exchange:", err)
	}
}

func PublicarEvento(cola string, payload []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := Canal.PublishWithContext(ctx,
		Exchange,    // Exchange
		cola,  // Routing key
		false, // Mandatory
		false, // Immediate
			amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         payload,
		})
	return err
}
