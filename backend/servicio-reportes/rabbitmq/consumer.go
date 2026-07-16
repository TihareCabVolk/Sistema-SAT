package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"servicio-reportes/models"
	"servicio-reportes/repository"

	amqp "github.com/rabbitmq/amqp091-go"
)

const dlxName = "sat.events.dlq"

func StartConsumer(ctx context.Context) {
	ch, err := Conn.Channel()
	if err != nil {
		log.Printf("[consumer] error creando canal: %v", err)
		return
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(dlxName, "topic", true, false, false, false, nil); err != nil {
		log.Printf("[consumer] error declarando DLX: %v", err)
		return
	}

	dlq, err := ch.QueueDeclare("reportes.alertas.dlq", true, false, false, false, nil)
	if err != nil {
		log.Printf("[consumer] error declarando DLQ: %v", err)
		return
	}
	if err := ch.QueueBind(dlq.Name, "#", dlxName, false, nil); err != nil {
		log.Printf("[consumer] error bindeando DLQ: %v", err)
		return
	}

	args := amqp.Table{"x-dead-letter-exchange": dlxName}
	q, err := ch.QueueDeclare("reportes.alertas", true, false, false, false, args)
	if err != nil {
		log.Printf("[consumer] error declarando cola: %v", err)
		return
	}

	if err := ch.QueueBind(q.Name, "alerta_emitida", Exchange, false, nil); err != nil {
		log.Printf("[consumer] error bindeando cola: %v", err)
		return
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("[consumer] error consumiendo: %v", err)
		return
	}

	log.Println("[consumer] esperando eventos alerta_emitida...")

	for {
		select {
		case <-ctx.Done():
			log.Println("[consumer] contexto cancelado, cerrando...")
			return
		case d, ok := <-msgs:
			if !ok {
				return
			}

			var alerta models.AlertaEmitida
			if err := json.Unmarshal(d.Body, &alerta); err != nil {
				log.Printf("[consumer] error parseando mensaje: %v", err)
				d.Nack(false, false)
				continue
			}

			estado, err := repository.EstadoPorTraceID(ctx, alerta.IdValidacion)
			if err != nil {
				log.Printf("[consumer] error verificando estado: %v", err)
				d.Nack(false, true)
				continue
			}
			if estado == "EMITIDA" {
				d.Ack(false)
				continue
			}

			log.Printf("[consumer] alerta recibida: id_validacion=%s", alerta.IdValidacion)

			if err := repository.ActualizarEstado(alerta.IdValidacion, "EMITIDA"); err != nil {
				log.Printf("[consumer] error actualizando estado: %v", err)
				d.Nack(false, true)
				continue
			}

			d.Ack(false)
			fmt.Printf("[SERVICIO-REPORTES] Registro %s actualizado a estado EMITIDA\n", alerta.IdValidacion)
		}
	}
}
