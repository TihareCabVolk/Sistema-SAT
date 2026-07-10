package models

import "time"

// contrato de datos de la alerta emitida
// cada campo debe ser igual al de logistica, sino se ignorara y se perdera el dato
type AlertaEmitida struct {
	Evento          string    `json:"evento"`
	Timestamp       time.Time `json:"timestamp"`
	IdValidacion    string    `json:"id_validacion"`
	NivelAlerta     string    `json:"nivel_alerta"`
	ZonasAfectadas  []string  `json:"zonas_afectadas"`
	CostoEmergencia float64   `json:"costo_emergencia"`
	Estado          string    `json:"estado"`
}