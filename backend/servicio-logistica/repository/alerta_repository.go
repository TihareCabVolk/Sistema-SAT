package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"servicio-logistica/models"
)

type AlertaRepository struct {
	db *sql.DB
}

func NewAlertaRepository(db *sql.DB) *AlertaRepository {
	return &AlertaRepository{db: db}
}

func (r *AlertaRepository) Insertar(ctx context.Context, a *models.Alerta) error {
	zonasJSON, _ := json.Marshal(a.ZonasAfectadas)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO alertas (id, id_validacion, magnitud, epicentro_lat, epicentro_lon,
		                     nivel_alerta, costo_emergencia, zonas_afectadas, estado, creado_en)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
	`, a.ID, a.IdValidacion, a.Magnitud, a.EpicentroLat, a.EpicentroLon,
		a.NivelAlerta, a.CostoEmergencia, zonasJSON, a.Estado)
	return err
}

func (r *AlertaRepository) Listar(ctx context.Context) ([]models.Alerta, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, id_validacion, magnitud, epicentro_lat, epicentro_lon,
		       nivel_alerta, costo_emergencia, zonas_afectadas, estado, creado_en
		FROM alertas ORDER BY creado_en DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	alertas := make([]models.Alerta, 0)
	for rows.Next() {
		var a models.Alerta
		var zonasJSON []byte
		rows.Scan(&a.ID, &a.IdValidacion, &a.Magnitud,
			&a.EpicentroLat, &a.EpicentroLon, &a.NivelAlerta,
			&a.CostoEmergencia, &zonasJSON, &a.Estado, &a.CreadoEn)
		json.Unmarshal(zonasJSON, &a.ZonasAfectadas)
		alertas = append(alertas, a)
	}
	return alertas, rows.Err()
}

func (r *AlertaRepository) ObtenerPorID(ctx context.Context, id string) (*models.Alerta, error) {
	var a models.Alerta
	var zonasJSON []byte
	err := r.db.QueryRowContext(ctx, `
		SELECT id, id_validacion, magnitud, epicentro_lat, epicentro_lon,
		       nivel_alerta, costo_emergencia, zonas_afectadas, estado, creado_en
		FROM alertas WHERE id = $1
	`, id).Scan(&a.ID, &a.IdValidacion, &a.Magnitud,
		&a.EpicentroLat, &a.EpicentroLon, &a.NivelAlerta,
		&a.CostoEmergencia, &zonasJSON, &a.Estado, &a.CreadoEn)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(zonasJSON, &a.ZonasAfectadas)
	return &a, nil
}

func (r *AlertaRepository) ExistePorIdValidacion(ctx context.Context, idValidacion string) (bool, error) {
	var existe bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM alertas WHERE id_validacion = $1)`, idValidacion).Scan(&existe)
	return existe, err
}
