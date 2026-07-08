package publisher

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"servicio-validacion/models"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitPublisher struct {
	conn     *amqp.Connection
	exchange string
}

func NewRabbitPublisher(conn *amqp.Connection, exchange string) *RabbitPublisher {
	return &RabbitPublisher{conn: conn, exchange: exchange}
}

func (p *RabbitPublisher) PublicarValidacion(ctx context.Context, vp *models.ValidacionPositiva) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(p.exchange, "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	body, err := json.Marshal(vp)
	if err != nil {
		return err
	}

	ctxPub, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctxPub,
		p.exchange,
		"validacion_positiva",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		return err
	}

	log.Printf("[publisher] validacion_positiva publicada: id_senal=%s", vp.IdSenal)
	return nil
}
