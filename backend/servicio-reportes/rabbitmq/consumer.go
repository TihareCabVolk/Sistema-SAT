package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"

	"servicio-reportes/models"
	"servicio-reportes/repository"

	amqp "github.com/rabbitmq/amqp091-go"
)

// El consumer que escucha la alerta_emitida, actualiza estado = 'EMITIDA'
func StartConsumer() {
	ch, err := Conn.Channel()
	if err != nil {
		log.Fatalf("Error creando canal consumer: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("reportes.alertas", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Error declarando cola: %v", err)
	}

	err = ch.QueueBind(q.Name, "alerta_emitida", Exchange, false, nil)
	if err != nil {
		log.Fatalf("Error bindeando cola: %v", err)
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Error consumiendo: %v", err)
	}

	log.Println("[consumer] esperando eventos alerta_emitida...")

	for d := range msgs {
		var alerta models.AlertaEmitida
		if err := json.Unmarshal(d.Body, &alerta); err != nil {
			log.Printf("[consumer] error parseando mensaje: %v", err)
			continue
		}

		log.Printf("[consumer] alerta recibida: id_validacion=%s", alerta.IdValidacion)

		if err := repository.ActualizarEstado(alerta.IdValidacion, "EMITIDA"); err != nil {
			log.Printf("[consumer] error actualizando estado: %v", err)
			continue
		}

		fmt.Printf("[SERVICIO-REPORTES] Registro %s actualizado a estado EMITIDA\n", alerta.IdValidacion)
	}
}