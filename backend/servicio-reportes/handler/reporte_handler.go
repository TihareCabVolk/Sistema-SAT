package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"servicio-reportes/models"
	"servicio-reportes/rabbitmq"
	"servicio-reportes/repository"

	"github.com/google/uuid"
)

func HandleReport(w http.ResponseWriter, r *http.Request) {
	var report models.SensorReport

	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		http.Error(w, "Estructura JSON inválida", http.StatusBadRequest)
		return
	}

	traceID := uuid.New().String()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":   "recibido",
		"trace_id": traceID,
		"message":  "Señal capturada. Procesando guardado asíncrono.",
	}); err != nil {
		fmt.Printf("[TRACE-%s] error escribiendo respuesta: %v\n", traceID, err)
	}

	go func(id string, data models.SensorReport) {
		fmt.Printf("[TRACE-%s] Guardando registro en Postgres\n", id)

		err := repository.GuardarSismo(id, data)
		if err != nil {
			fmt.Printf("[TRACE-%s] ERROR EN BD: %v\n", id, err)
			return
		}

		fmt.Printf("[TRACE-%s] Guardado en DB_Reportes\n", id)

		eventoRabbit := models.SenalRecibidaEvent{
			Evento:        "señal_recibida",
			IDSenal:       id,
			Timestamp:     data.Timestamp,
			IDSensor:      data.IDSensor,
			Ubicacion:     data.Ubicacion,
			Magnitud:      data.Magnitud,
			ProfundidadKm: data.ProfundidadKm,
			Confianza:     data.Confianza,
		}

		payloadJSON, err := json.Marshal(eventoRabbit)
		if err != nil {
			fmt.Printf("[TRACE-%s] Error al serializar JSON para Rabbit: %v\n", id, err)
			return
		}

		err = rabbitmq.PublicarEvento("señal_recibida", payloadJSON)
		if err != nil {
			fmt.Printf("[TRACE-%s] ERROR EN RABBITMQ: %v\n", id, err)
			return
		}

		fmt.Printf("[TRACE-%s] Evento publicado en RabbitMQ correctamente %s\n", id, string(payloadJSON))

	}(traceID, report)
}
