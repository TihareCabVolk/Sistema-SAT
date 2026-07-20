package service

import (
	"context"
	"log"
	"time"

	"servicio-logistica/models"
	"servicio-logistica/repository"

	"github.com/google/uuid"
)

type LogisticaService struct {
	repo *repository.AlertaRepository
}

func NewLogisticaService(repo *repository.AlertaRepository) *LogisticaService {
	return &LogisticaService{repo: repo}
}

func (s *LogisticaService) ProcesarValidacion(ctx context.Context, vp *models.ValidacionPositiva) (*models.AlertaEmitida, error) {
	existe, err := s.repo.ExistePorIdValidacion(ctx, vp.IdSenal)
	if err != nil {
		return nil, err
	}
	if existe {
		log.Printf("[logistica] idempotencia: validacion %s ya procesada, ignorando", vp.IdSenal)
		return nil, nil
	}

	nivel := s.calcularNivelAlerta(vp.MagnitudFinal)
	costo := s.calcularCosto(vp.MagnitudFinal)
	zonas := vp.ZonasAfectadas
	if zonas == nil {
		zonas = []string{}
	}

	id := uuid.New().String()
	alerta := models.Alerta{
		ID:              id,
		IdValidacion:    vp.IdSenal,
		Magnitud:        vp.MagnitudFinal,
		EpicentroLat:    vp.Epicentro.Lat,
		EpicentroLon:    vp.Epicentro.Lon,
		NivelAlerta:     nivel,
		CostoEmergencia: costo,
		ZonasAfectadas:  zonas,
		Estado:          "EMITIDA",
		CreadoEn:        time.Now(),
	}

	if err := s.repo.Insertar(ctx, &alerta); err != nil {
		return nil, err
	}

	log.Printf("[logistica] alerta emitida: id=%s, nivel=%s, costo=%.0f", id, nivel, costo)

	return &models.AlertaEmitida{
		Evento:          "alerta_emitida",
		Timestamp:       time.Now().UTC(),
		IdValidacion:    vp.IdSenal,
		NivelAlerta:     nivel,
		ZonasAfectadas:  zonas,
		CostoEmergencia: costo,
		Estado:          "EMITIDA",
	}, nil
}

func (s *LogisticaService) ListarAlertas(ctx context.Context) ([]models.Alerta, error) {
	return s.repo.Listar(ctx)
}

func (s *LogisticaService) ObtenerAlerta(ctx context.Context, id string) (*models.Alerta, error) {
	return s.repo.ObtenerPorID(ctx, id)
}

func (s *LogisticaService) calcularNivelAlerta(magnitud float64) string {
	switch {
	case magnitud >= 6.0:
		return "ROJO"
	case magnitud >= 4.0:
		return "AMARILLO"
	default:
		return "VERDE"
	}
}

func (s *LogisticaService) calcularCosto(magnitud float64) float64 {
	switch {
	case magnitud >= 6.0:
		return 1500000
	case magnitud >= 4.0:
		return 150000
	default:
		return 50000
	}
}
