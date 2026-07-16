package consumer

import (
	"context"
	"encoding/json"
	"log"

	"servicio-logistica/models"
	"servicio-logistica/publisher"
	"servicio-logistica/service"

	amqp "github.com/rabbitmq/amqp091-go"
)

const dlxName = "sat.events.dlq"

type RabbitConsumer struct {
	conn      *amqp.Connection
	exchange  string
	svc       *service.LogisticaService
	publisher *publisher.RabbitPublisher
}

func NewRabbitConsumer(conn *amqp.Connection, exchange string, svc *service.LogisticaService, pub *publisher.RabbitPublisher) *RabbitConsumer {
	return &RabbitConsumer{
		conn:      conn,
		exchange:  exchange,
		svc:       svc,
		publisher: pub,
	}
}

func (c *RabbitConsumer) Start(ctx context.Context) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(dlxName, "topic", true, false, false, false, nil); err != nil {
		return err
	}

	dlq, err := ch.QueueDeclare("logistica.validaciones.dlq", true, false, false, false, nil)
	if err != nil {
		return err
	}
	if err := ch.QueueBind(dlq.Name, "#", dlxName, false, nil); err != nil {
		return err
	}

	if err := ch.ExchangeDeclare(c.exchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}

	args := amqp.Table{"x-dead-letter-exchange": dlxName}
	q, err := ch.QueueDeclare("logistica.validaciones", true, false, false, false, args)
	if err != nil {
		return err
	}

	if err := ch.QueueBind(q.Name, "validacion_positiva", c.exchange, false, nil); err != nil {
		return err
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	log.Println("[consumer] esperando eventos validacion_positiva...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-msgs:
			if !ok {
				return nil
			}

			var vp models.ValidacionPositiva
			if err := json.Unmarshal(d.Body, &vp); err != nil {
				log.Printf("[consumer] error parseando mensaje: %v", err)
				d.Nack(false, false)
				continue
			}

			existe, err := c.svc.ExistePorValidacion(ctx, vp.IdSenal)
			if err != nil {
				log.Printf("[consumer] error verificando idempotencia: %v", err)
				d.Nack(false, true)
				continue
			}
			if existe {
				d.Ack(false)
				continue
			}

			log.Printf("[consumer] validacion recibida: id_senal=%s", vp.IdSenal)

			alerta, err := c.svc.ProcesarValidacion(ctx, &vp)
			if err != nil {
				log.Printf("[consumer] error procesando validacion: %v", err)
				d.Nack(false, true)
				continue
			}

			if err := c.publisher.PublicarAlerta(ctx, alerta); err != nil {
				log.Printf("[consumer] error publicando alerta: %v", err)
				d.Nack(false, true)
				continue
			}

			d.Ack(false)
		}
	}
}
