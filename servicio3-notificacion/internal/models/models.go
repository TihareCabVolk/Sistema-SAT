package models

import "time"

// Envelope genérico de todos los eventos del sistema SAT.
// Mantener el mismo "sobre" en los 3 servicios facilita la trazabilidad
// (el event_id permite seguir un mensaje a través de todo el flujo).
type Event struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   map[string]any  `json:"-"` // se define por tipo concreto abajo
}

// SismoValidado es el payload que publica el Servicio 2.
type SismoValidado struct {
	ReporteID            string  `json:"reporte_id"`
	SensorID             string  `json:"sensor_id"`
	Magnitud             float64 `json:"magnitud"`
	Latitud              float64 `json:"latitud"`
	Longitud             float64 `json:"longitud"`
	ProfundidadKm        float64 `json:"profundidad_km"`
	SensoresConfirmantes int     `json:"sensores_confirmantes"`
}

// EventoSismoValidado = sobre + payload concreto (lo que consumimos).
type EventoSismoValidado struct {
	EventID   string        `json:"event_id"`
	EventType string        `json:"event_type"`
	Timestamp time.Time     `json:"timestamp"`
	Payload   SismoValidado `json:"payload"`
}

// AlertaEmitida es el payload que ESTE servicio publica al cerrar el proceso.
type AlertaEmitida struct {
	ReporteID          string   `json:"reporte_id"`
	AlertaID           string   `json:"alerta_id"`
	Nivel              string   `json:"nivel"`
	CostoEmergenciaCLP int64    `json:"costo_emergencia_clp"`
	CanalesNotificados []string `json:"canales_notificados"`
	Estado             string   `json:"estado"`
}

type EventoAlertaEmitida struct {
	EventID   string        `json:"event_id"`
	EventType string        `json:"event_type"`
	Timestamp time.Time     `json:"timestamp"`
	Payload   AlertaEmitida `json:"payload"`
}

// Alerta es la entidad persistida en el historial oficial (BD propia).
type Alerta struct {
	ID                 string
	ReporteID          string
	EventIDOrigen      string // event_id del sismo.validado (clave de idempotencia)
	Magnitud           float64
	Nivel              string
	CostoEmergenciaCLP int64
	Estado             string
	CreadaEn           time.Time
}
