package models

import "time"

type Ubicacion struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type ValidacionPositiva struct {
	Evento              string    `json:"evento"`
	Timestamp           time.Time `json:"timestamp"`
	IdSenal             string    `json:"id_senal"`
	SensoresConfirmados []string  `json:"sensores_confirmados"`
	MagnitudFinal       float64   `json:"magnitud_final"`
	Epicentro           Ubicacion `json:"epicentro"`
	ZonasAfectadas      []string  `json:"zonas_afectadas,omitempty"`
}

type AlertaEmitida struct {
	Evento          string    `json:"evento"`
	Timestamp       time.Time `json:"timestamp"`
	IdValidacion    string    `json:"id_validacion"`
	NivelAlerta     string    `json:"nivel_alerta"`
	ZonasAfectadas  []string  `json:"zonas_afectadas"`
	CostoEmergencia float64   `json:"costo_emergencia"`
	Estado          string    `json:"estado"`
}
