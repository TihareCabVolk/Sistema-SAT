package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/kundev/servicio3-notificacion/internal/models"
)

// ErrDuplicado indica que este evento ya fue procesado antes (idempotencia).
// Ocurre si RabbitMQ re-entrega un mensaje (ej: el pod murió antes del ACK).
var ErrDuplicado = errors.New("evento ya procesado")

type AlertaRepository struct {
	db *sql.DB
}

func NewAlertaRepository(db *sql.DB) *AlertaRepository {
	return &AlertaRepository{db: db}
}

// Guardar inserta la alerta en el historial oficial.
// La restricción UNIQUE sobre event_id_origen garantiza idempotencia:
// si el mismo evento llega dos veces, la segunda inserción falla con 23505.
func (r *AlertaRepository) Guardar(ctx context.Context, a models.Alerta) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO alertas (id, reporte_id, event_id_origen, magnitud, nivel, costo_emergencia_clp, estado, creada_en)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		a.ID, a.ReporteID, a.EventIDOrigen, a.Magnitud, a.Nivel, a.CostoEmergenciaCLP, a.Estado, a.CreadaEn,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return ErrDuplicado
		}
		return fmt.Errorf("insertando alerta: %w", err)
	}
	return nil
}

// BuscarPorReporte es útil para debugging y para el manual operativo.
func (r *AlertaRepository) BuscarPorReporte(ctx context.Context, reporteID string) (*models.Alerta, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, reporte_id, event_id_origen, magnitud, nivel, costo_emergencia_clp, estado, creada_en
		FROM alertas WHERE reporte_id = $1`, reporteID)

	var a models.Alerta
	err := row.Scan(&a.ID, &a.ReporteID, &a.EventIDOrigen, &a.Magnitud, &a.Nivel, &a.CostoEmergenciaCLP, &a.Estado, &a.CreadaEn)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}
