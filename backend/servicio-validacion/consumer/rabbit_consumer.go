package consumer

import (
	"context"
	"encoding/json"
	"log"

	"servicio-validacion/models"
	"servicio-validacion/publisher"
	"servicio-validacion/service"

	amqp "github.com/rabbitmq/amqp091-go"
)

const dlxName = "sat.events.dlq"

type RabbitConsumer struct {
	conn      *amqp.Connection
	cola      string
	exchange  string
	svc       *service.ValidacionService
	publisher *publisher.RabbitPublisher
}

func NewRabbitConsumer(conn *amqp.Connection, cola string, exchange string,
	svc *service.ValidacionService, pub *publisher.RabbitPublisher) *RabbitConsumer {
	return &RabbitConsumer{
		conn:      conn,
		cola:      cola,
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

	dlq, err := ch.QueueDeclare(c.cola+".dlq", true, false, false, false, nil)
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
	q, err := ch.QueueDeclare(c.cola, true, false, false, false, args)
	if err != nil {
		return err
	}

	if err := ch.QueueBind(q.Name, "señal_recibida", c.exchange, false, nil); err != nil {
		return err
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	log.Println("[consumer] esperando eventos señal_recibida...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-msgs:
			if !ok {
				return nil
			}

			var sr models.SenalRecibida
			if err := json.Unmarshal(d.Body, &sr); err != nil {
				log.Printf("[consumer] error parseando mensaje: %v", err)
				d.Nack(false, false)
				continue
			}

			valida, err := c.svc.EstaValidada(ctx, sr.IDSenal)
			if err != nil {
				log.Printf("[consumer] error verificando idempotencia: %v", err)
				d.Nack(false, true)
				continue
			}
			if valida {
				d.Ack(false)
				continue
			}

			log.Printf("[consumer] señal recibida: id_sensor=%s", sr.IDSensor)

			validacion, err := c.svc.ProcesarSenal(ctx, &sr)
			if err != nil {
				log.Printf("[consumer] error procesando señal: %v", err)
				d.Nack(false, true)
				continue
			}

			if validacion == nil {
				d.Ack(false)
				continue
			}

			if err := c.publisher.PublicarValidacion(ctx, validacion); err != nil {
				log.Printf("[consumer] error publicando validacion: %v", err)
				d.Nack(false, true)
				continue
			}

			d.Ack(false)
		}
	}
}
