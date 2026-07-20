package repository

import (
	"context"
	"database/sql"

	"servicio-validacion/models"

	"github.com/lib/pq"
)

type SenalRepository struct {
	db *sql.DB
}

func NewSenalRepository(db *sql.DB) *SenalRepository {
	return &SenalRepository{db: db}
}

func (r *SenalRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *SenalRepository) Insertar(ctx context.Context, tx *sql.Tx, s *models.Senal) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO senales (id, id_sensor, lat, lon, magnitud, profundidad_km, confianza, timestamp, validada)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO NOTHING
	`, s.ID, s.IDSensor, s.Lat, s.Lon, s.Magnitud, s.ProfundidadKm, s.Confianza, s.Timestamp, s.Validada)
	return err
}

// BuscarRecientesParaActualizar bloquea (FOR UPDATE) las señales candidatas dentro de la
// ventana de tiempo para que ninguna otra réplica pueda leer/marcar el mismo grupo hasta
// que esta transacción termine (commit o rollback).
func (r *SenalRepository) BuscarRecientesParaActualizar(ctx context.Context, tx *sql.Tx, ventanaSeg int) ([]models.Senal, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, id_sensor, lat, lon, magnitud, profundidad_km, confianza, timestamp, validada
		FROM senales
		WHERE validada = false AND timestamp >= NOW() - ($1::text || ' seconds')::interval
		ORDER BY id
		FOR UPDATE
	`, ventanaSeg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var senales []models.Senal
	for rows.Next() {
		var s models.Senal
		if err := rows.Scan(&s.ID, &s.IDSensor, &s.Lat, &s.Lon, &s.Magnitud,
			&s.ProfundidadKm, &s.Confianza, &s.Timestamp, &s.Validada); err != nil {
			return nil, err
		}
		senales = append(senales, s)
	}
	return senales, rows.Err()
}

func (r *SenalRepository) MarcarValidadas(ctx context.Context, tx *sql.Tx, ids []string) error {
	_, err := tx.ExecContext(ctx, `UPDATE senales SET validada = true WHERE id = ANY($1)`, pq.Array(ids))
	return err
}

func (r *SenalRepository) EstaValidada(ctx context.Context, id string) (bool, error) {
	var validada bool
	err := r.db.QueryRowContext(ctx, `SELECT validada FROM senales WHERE id = $1`, id).Scan(&validada)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return validada, err
}
