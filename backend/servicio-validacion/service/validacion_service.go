package service

import (
	"context"
	"log"
	"math"
	"time"

	"servicio-validacion/models"
	"servicio-validacion/repository"
)

type ValidacionService struct {
	repo        *repository.SenalRepository
	radioKm     float64
	ventanaSeg  int
	minSensores int
}

func NewValidacionService(repo *repository.SenalRepository, radioKm float64, ventanaSeg int, minSensores int) *ValidacionService {
	return &ValidacionService{
		repo:        repo,
		radioKm:     radioKm,
		ventanaSeg:  ventanaSeg,
		minSensores: minSensores,
	}
}

func (s *ValidacionService) ProcesarSenal(ctx context.Context, sr *models.SenalRecibida) (*models.ValidacionPositiva, error) {
	validada, err := s.repo.EstaValidada(ctx, sr.IDSenal)
	if err != nil {
		return nil, err
	}
	if validada {
		log.Printf("[validacion] idempotencia: señal %s ya validada, ignorando", sr.IDSenal)
		return nil, nil
	}

	senal := &models.Senal{
		ID:            sr.IDSenal,
		IDSensor:      sr.IDSensor,
		Lat:           sr.Ubicacion.Lat,
		Lon:           sr.Ubicacion.Lon,
		Magnitud:      sr.Magnitud,
		ProfundidadKm: sr.ProfundidadKm,
		Confianza:     sr.Confianza,
		Timestamp:     sr.Timestamp,
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := s.repo.Insertar(ctx, tx, senal); err != nil {
		return nil, err
	}

	// FOR UPDATE bloquea las filas candidatas hasta el commit: si otra réplica está
	// procesando una señal que cae en el mismo grupo, espera aquí en vez de contar
	// sensores en paralelo sobre datos que todavía no se marcaron como validados.
	recientes, err := s.repo.BuscarRecientesParaActualizar(ctx, tx, s.ventanaSeg)
	if err != nil {
		return nil, err
	}

	cercanas := s.filtrarCercanas(senal, recientes)

	sensores := s.sensoresUnicos(cercanas)
	if len(sensores) < s.minSensores {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		log.Printf("[validacion] señal %s guardada, %d/%d sensores confirmados", senal.ID, len(sensores), s.minSensores)
		return nil, nil
	}

	var ids []string
	var magnitudTotal, latTotal, lonTotal float64
	for _, c := range cercanas {
		ids = append(ids, c.ID)
		magnitudTotal += c.Magnitud
		latTotal += c.Lat
		lonTotal += c.Lon
	}
	n := float64(len(cercanas))

	// Marca el grupo como procesado dentro de la misma transacción: al hacer commit,
	// ninguna otra réplica volverá a ver estas filas como validada=false, así que no
	// puede volver a publicar validacion_positiva para el mismo grupo.
	if err := s.repo.MarcarValidadas(ctx, tx, ids); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	validacion := &models.ValidacionPositiva{
		Evento:              "validacion_positiva",
		Timestamp:           time.Now().UTC(),
		IdSenal:             senal.ID,
		SensoresConfirmados: sensores,
		MagnitudFinal:       magnitudTotal / n,
		Epicentro: models.Ubicacion{
			Lat: latTotal / n,
			Lon: lonTotal / n,
		},
	}

	log.Printf("[validacion] sismo confirmado: id_senal=%s sensores=%v", validacion.IdSenal, sensores)

	return validacion, nil
}

func (s *ValidacionService) filtrarCercanas(origen *models.Senal, candidatas []models.Senal) []models.Senal {
	var cercanas []models.Senal
	for _, c := range candidatas {
		if distanciaKm(origen.Lat, origen.Lon, c.Lat, c.Lon) <= s.radioKm {
			cercanas = append(cercanas, c)
		}
	}
	return cercanas
}

func (s *ValidacionService) sensoresUnicos(senales []models.Senal) []string {
	vistos := make(map[string]bool)
	var sensores []string
	for _, sn := range senales {
		if !vistos[sn.IDSensor] {
			vistos[sn.IDSensor] = true
			sensores = append(sensores, sn.IDSensor)
		}
	}
	return sensores
}

// formula de haversine
func distanciaKm(lat1, lon1, lat2, lon2 float64) float64 {
	const radioTierraKm = 6371
	dLat := gradosARadianes(lat2 - lat1)
	dLon := gradosARadianes(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(gradosARadianes(lat1))*math.Cos(gradosARadianes(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return radioTierraKm * c
}

func gradosARadianes(g float64) float64 {
	return g * math.Pi / 180
}
