// Simulador del Servicio 2 (Validación de Sensores).
// Publica un evento "sismo.validado" idéntico al que publicará el servicio real.
// Cuando el Servicio 2 exista, este archivo simplemente deja de usarse:
// tu Servicio 3 no nota ninguna diferencia porque el contrato es el mismo.
//
// Uso:
//   go run ./tools/simulador-servicio2 -magnitud 7.2
//   go run ./tools/simulador-servicio2 -magnitud 4.8 -sensor SENSOR-IQUIQUE-03
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/kundev/servicio3-notificacion/internal/models"
)

func main() {
	magnitud := flag.Float64("magnitud", 6.8, "magnitud del sismo simulado")
	sensor := flag.String("sensor", "SENSOR-ARICA-01", "id del sensor")
	flag.Parse()

	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("conectando a RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	ev := models.EventoSismoValidado{
		EventID:   uuid.NewString(),
		EventType: "sismo.validado",
		Timestamp: time.Now().UTC(),
		Payload: models.SismoValidado{
			ReporteID:            uuid.NewString(),
			SensorID:             *sensor,
			Magnitud:             *magnitud,
			Latitud:              -18.4783,
			Longitud:             -70.3126,
			ProfundidadKm:        35.2,
			SensoresConfirmantes: 3,
		},
	}

	body, _ := json.MarshalIndent(ev, "", "  ")
	err = ch.PublishWithContext(context.Background(),
		"sat.eventos",     // mismo exchange que usará el Servicio 2 real
		"sismo.validado",  // misma routing key
		false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    ev.EventID,
			Body:         body,
		})
	if err != nil {
		log.Fatalf("publicando: %v", err)
	}

	log.Printf("evento publicado:\n%s", body)
}
