package models

import "time"

type Ubicacion struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// lo que manda servicio 1
type SenalRecibida struct {
	Evento        string    `json:"evento"`
	Timestamp     time.Time `json:"timestamp"`
	IDSensor      string    `json:"id_sensor"`
	Ubicacion     Ubicacion `json:"ubicacion"`
	Magnitud      float64   `json:"magnitud"`
	ProfundidadKm int       `json:"profundidad_km"`
	Confianza     float64   `json:"confianza"`
}

// lo que le mandamos a servicio 3
type ValidacionPositiva struct {
	Evento              string    `json:"evento"`
	Timestamp           time.Time `json:"timestamp"`
	IdSenal             string    `json:"id_senal"`
	SensoresConfirmados []string  `json:"sensores_confirmados"`
	MagnitudFinal       float64   `json:"magnitud_final"`
	Epicentro           Ubicacion `json:"epicentro"`
}

// fila guardada en la bd propia
type Senal struct {
	ID            string
	IDSensor      string
	Lat           float64
	Lon           float64
	Magnitud      float64
	ProfundidadKm int
	Confianza     float64
	Timestamp     time.Time
	Validada      bool
}
