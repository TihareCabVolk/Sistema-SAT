package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/kundev/servicio3-notificacion/internal/models"
)

const (
	Exchange       = "sat.eventos"         // exchange topic compartido por los 3 servicios
	QueueName      = "q.servicio3.sismo-validado"
	BindingKey     = "sismo.validado"      // lo que consumimos
	PublishKey     = "alerta.emitida"      // lo que publicamos
	DLXExchange    = "sat.eventos.dlx"     // dead letter para mensajes venenosos
	DLQName        = "q.servicio3.dlq"
)

type RabbitMQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// Connect abre conexión con reintentos (importante en Kubernetes:
// el pod del servicio puede levantar antes que el pod de RabbitMQ).
func Connect(url string) (*RabbitMQ, error) {
	var conn *amqp.Connection
	var err error
	for i := 1; i <= 10; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("[broker] intento %d/10 fallido: %v — reintentando en 3s", i, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("conectando a RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	r := &RabbitMQ{conn: conn, ch: ch}
	if err := r.declararTopologia(); err != nil {
		return nil, err
	}
	return r, nil
}

// declararTopologia crea exchange, cola, DLQ y bindings.
// Es idempotente: si ya existen con los mismos parámetros, no pasa nada.
func (r *RabbitMQ) declararTopologia() error {
	// Exchange principal (topic: permite routing por patrones como "sismo.*")
	if err := r.ch.ExchangeDeclare(Exchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	// Dead Letter Exchange: adonde van los mensajes rechazados sin requeue
	if err := r.ch.ExchangeDeclare(DLXExchange, "fanout", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := r.ch.QueueDeclare(DLQName, true, false, false, false, nil); err != nil {
		return err
	}
	if err := r.ch.QueueBind(DLQName, "", DLXExchange, false, nil); err != nil {
		return err
	}
	// Cola principal, durable, con DLX configurado
	args := amqp.Table{"x-dead-letter-exchange": DLXExchange}
	if _, err := r.ch.QueueDeclare(QueueName, true, false, false, false, args); err != nil {
		return err
	}
	if err := r.ch.QueueBind(QueueName, BindingKey, Exchange, false, nil); err != nil {
		return err
	}
	// Prefetch=1: no recibir un nuevo mensaje hasta terminar (ACK) el actual.
	// Evita que un pod acapare mensajes que no alcanza a procesar.
	return r.ch.Qos(1, 0, false)
}

// Handler es la función que procesa cada evento consumido.
type Handler func(ctx context.Context, ev models.EventoSismoValidado) error

// Consumir inicia el loop de consumo. Bloquea hasta que ctx se cancele.
func (r *RabbitMQ) Consumir(ctx context.Context, handle Handler) error {
	msgs, err := r.ch.Consume(QueueName, "servicio3", false /* autoAck=false */, false, false, false, nil)
	if err != nil {
		return err
	}
	log.Printf("[broker] escuchando cola %s (binding %s)", QueueName, BindingKey)

	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-msgs:
			if !ok {
				return fmt.Errorf("canal de consumo cerrado")
			}
			r.procesarEntrega(ctx, d, handle)
		}
	}
}

func (r *RabbitMQ) procesarEntrega(ctx context.Context, d amqp.Delivery, handle Handler) {
	var ev models.EventoSismoValidado
	if err := json.Unmarshal(d.Body, &ev); err != nil {
		// Mensaje malformado: reintentarlo no lo arreglará ("mensaje venenoso").
		// Nack SIN requeue -> se va a la Dead Letter Queue para análisis.
		log.Printf("[broker] JSON inválido, enviando a DLQ: %v", err)
		_ = d.Nack(false, false)
		return
	}

	if err := handle(ctx, ev); err != nil {
		// Error transitorio (ej: BD caída) -> Nack CON requeue para reintentar.
		log.Printf("[broker] error procesando %s: %v — requeue", ev.EventID, err)
		_ = d.Nack(false, true)
		return
	}

	// Solo confirmamos (ACK) cuando la alerta quedó persistida Y publicada.
	_ = d.Ack(false)
}

// PublicarAlertaEmitida implementa la interfaz service.Publisher.
func (r *RabbitMQ) PublicarAlertaEmitida(ctx context.Context, ev models.EventoAlertaEmitida) error {
	body, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	return r.ch.PublishWithContext(ctx, Exchange, PublishKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent, // sobrevive reinicios del broker
		MessageId:    ev.EventID,
		Timestamp:    ev.Timestamp,
		Body:         body,
	})
}

func (r *RabbitMQ) Close() {
	if r.ch != nil {
		_ = r.ch.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
}
