package publisher

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"servicio-logistica/models"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitPublisher struct {
	conn     *amqp.Connection
	exchange string
}

func NewRabbitPublisher(conn *amqp.Connection, exchange string) *RabbitPublisher {
	return &RabbitPublisher{conn: conn, exchange: exchange}
}

func (p *RabbitPublisher) PublicarAlerta(ctx context.Context, alerta *models.AlertaEmitida) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	body, err := json.Marshal(alerta)
	if err != nil {
		return err
	}

	ctxPub, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctxPub,
		p.exchange,
		"alerta_emitida",
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

	log.Printf("[publisher] alerta_emitida publicada: id=%s", alerta.IdValidacion)
	return nil
}
