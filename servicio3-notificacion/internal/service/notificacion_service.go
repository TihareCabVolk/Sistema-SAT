package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/kundev/servicio3-notificacion/internal/models"
	"github.com/kundev/servicio3-notificacion/internal/repository"
)

// Publisher es la interfaz que implementa el broker.
// Definirla aquí (y no importar el paquete broker) invierte la dependencia:
// el servicio de negocio no sabe que existe RabbitMQ. Esto se llama
// "Dependency Inversion" y te permite testear con un mock.
type Publisher interface {
	PublicarAlertaEmitida(ctx context.Context, ev models.EventoAlertaEmitida) error
}

type NotificacionService struct {
	repo      *repository.AlertaRepository
	publisher Publisher
}

func NewNotificacionService(repo *repository.AlertaRepository, pub Publisher) *NotificacionService {
	return &NotificacionService{repo: repo, publisher: pub}
}

// ProcesarSismoValidado es el corazón del Servicio 3:
// 1. Calcula nivel y costos de emergencia (protocolo)
// 2. Registra en el historial oficial (BD propia)
// 3. Publica alerta.emitida para cerrar el ciclo
func (s *NotificacionService) ProcesarSismoValidado(ctx context.Context, ev models.EventoSismoValidado) error {
	nivel := calcularNivel(ev.Payload.Magnitud)
	costo := calcularCostoEmergencia(nivel)

	alerta := models.Alerta{
		ID:                 uuid.NewString(),
		ReporteID:          ev.Payload.ReporteID,
		EventIDOrigen:      ev.EventID,
		Magnitud:           ev.Payload.Magnitud,
		Nivel:              nivel,
		CostoEmergenciaCLP: costo,
		Estado:             "EMITIDA",
		CreadaEn:           time.Now().UTC(),
	}

	// Paso 1: persistir PRIMERO. Si la BD falla, hacemos NACK y RabbitMQ
	// reintenta. Si publicáramos primero y la BD fallara, emitiríamos una
	// alerta sin registro oficial (inconsistencia).
	if err := s.repo.Guardar(ctx, alerta); err != nil {
		if errors.Is(err, repository.ErrDuplicado) {
			log.Printf("[service] evento %s ya procesado, se descarta (idempotencia)", ev.EventID)
			return nil // ACK: no queremos reprocesar duplicados
		}
		return err // NACK con requeue
	}

	// Paso 2: publicar el cierre del proceso
	salida := models.EventoAlertaEmitida{
		EventID:   uuid.NewString(),
		EventType: "alerta.emitida",
		Timestamp: time.Now().UTC(),
		Payload: models.AlertaEmitida{
			ReporteID:          alerta.ReporteID,
			AlertaID:           alerta.ID,
			Nivel:              nivel,
			CostoEmergenciaCLP: costo,
			CanalesNotificados: []string{"ciudadania", "equipos_emergencia", "autoridades"},
			Estado:             "EMITIDA",
		},
	}
	if err := s.publisher.PublicarAlertaEmitida(ctx, salida); err != nil {
		return err
	}

	log.Printf("[service] alerta %s emitida (reporte=%s nivel=%s costo=%d CLP)",
		alerta.ID, alerta.ReporteID, nivel, costo)
	return nil
}

// calcularNivel: reglas simples del protocolo de emergencia.
func calcularNivel(magnitud float64) string {
	switch {
	case magnitud >= 7.0:
		return "ROJA"
	case magnitud >= 5.5:
		return "NARANJA"
	default:
		return "AMARILLA"
	}
}

// calcularCostoEmergencia: "activa los protocolos de costos de emergencia".
// Valores de ejemplo; lo importante es que la lógica viva aquí, aislada.
func calcularCostoEmergencia(nivel string) int64 {
	switch nivel {
	case "ROJA":
		return 15_000_000
	case "NARANJA":
		return 5_000_000
	default:
		return 1_000_000
	}
}
