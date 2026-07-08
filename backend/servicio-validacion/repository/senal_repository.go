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

func (r *SenalRepository) Insertar(ctx context.Context, s *models.Senal) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO senales (id, id_sensor, lat, lon, magnitud, profundidad_km, confianza, timestamp, validada)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO NOTHING
	`, s.ID, s.IDSensor, s.Lat, s.Lon, s.Magnitud, s.ProfundidadKm, s.Confianza, s.Timestamp, s.Validada)
	return err
}

func (r *SenalRepository) BuscarRecientes(ctx context.Context, ventanaSeg int) ([]models.Senal, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, id_sensor, lat, lon, magnitud, profundidad_km, confianza, timestamp, validada
		FROM senales
		WHERE validada = false AND timestamp >= NOW() - ($1 || ' seconds')::interval
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

func (r *SenalRepository) MarcarValidadas(ctx context.Context, ids []string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE senales SET validada = true WHERE id = ANY($1)`, pq.Array(ids))
	return err
}
