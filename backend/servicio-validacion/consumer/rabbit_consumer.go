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

	// servicio 1 publica al exchange default, la cola queda accesible directo por su nombre
	q, err := ch.QueueDeclare(c.cola, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(q.Name, "señal_recibida", c.exchange, false, nil)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
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
				continue
			}

			log.Printf("[consumer] señal recibida: id_sensor=%s", sr.IDSensor)

			validacion, err := c.svc.ProcesarSenal(ctx, &sr)
			if err != nil {
				log.Printf("[consumer] error procesando señal: %v", err)
				continue
			}

			if validacion == nil {
				continue
			}

			if err := c.publisher.PublicarValidacion(ctx, validacion); err != nil {
				log.Printf("[consumer] error publicando validacion: %v", err)
			}
		}
	}
}
