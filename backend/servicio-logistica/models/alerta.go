package models

import "time"

type Alerta struct {
	ID              string    `json:"id"`
	IdValidacion    string    `json:"id_validacion"`
	Magnitud        float64   `json:"magnitud"`
	EpicentroLat    float64   `json:"epicentro_lat"`
	EpicentroLon    float64   `json:"epicentro_lon"`
	NivelAlerta     string    `json:"nivel_alerta"`
	CostoEmergencia float64   `json:"costo_emergencia"`
	ZonasAfectadas  []string  `json:"zonas_afectadas"`
	Estado          string    `json:"estado"`
	CreadoEn        time.Time `json:"creado_en"`
}
