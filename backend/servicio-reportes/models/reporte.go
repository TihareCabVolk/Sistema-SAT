package models

import "time"

// coordenadas geográficas
type Ubicacion struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// contrato de entrada
type SensorReport struct {
	IDSensor      string    `json:"id_sensor"`
	Ubicacion     Ubicacion `json:"ubicacion"`
	Magnitud      float64   `json:"magnitud"`
	ProfundidadKm int       `json:"profundidad_km"`
	Confianza     float64   `json:"confianza"`
	Timestamp     time.Time `json:"timestamp"`
}

//contrato de salida
type SenalRecibidaEvent struct {
	Evento        string    `json:"evento"`
	IDSenal       string    `json:"id_senal"`
	Timestamp     time.Time `json:"timestamp"`
	IDSensor      string    `json:"id_sensor"`
	Ubicacion     Ubicacion `json:"ubicacion"`
	Magnitud      float64   `json:"magnitud"`
	ProfundidadKm int       `json:"profundidad_km"`
	Confianza     float64   `json:"confianza"`
}
