# Sistema de Alerta Temprana de Sismos - SAT

## Descripción

Este proyecto implementa un sistema distribuido basado en una arquitectura de microservicios para la gestión de alertas tempranas de sismos.

La solución utiliza comunicación síncrona mediante REST para la recepción inicial de reportes y comunicación asíncrona basada en eventos para el procesamiento interno, permitiendo alta disponibilidad, desacoplamiento entre servicios y tolerancia a grandes volúmenes de eventos.

---

## Arquitectura del Sistema

| Componente | Tecnología | Puerto |
|-----------|------------|--------|
| Frontend | React + Nginx Alpine | 80 |
| API Gateway | Nginx Alpine | 80 |
| Servicio 1 - Reportes | Go | 4001 |
| Servicio 2 - Validación | Go | 4002 |
| Servicio 3 - Logística | Go + Gin | 4003 |
| DB Reportes | PostgreSQL 16 Alpine | 5432 |
| DB Validación | PostgreSQL 16 Alpine | 5432 |
| DB Logística | PostgreSQL 16 Alpine | 5432 |
| Message Broker | RabbitMQ 3.13 Alpine | 5672 |

---

## Bases de Datos

Se utilizan tres bases de datos independientes, que son las siguientes:

| Servicio | Base de Datos |
|----------|---------------|
| Centro de Reportes | DB Reportes |
| Validación | DB Historial Geográfico |
| Logística | DB Costos y Alertas |

---

## Contrato de Datos (Eventos RabbitMQ)

Todos los eventos viajan por el exchange `sat.events` (tipo `topic`).

### Evento: `senal_recibida`

Publicado por el **Servicio 1: Reportes** al recibir una señal de un sensor.

```json
{
  "evento": "senal_recibida",
  "id_senal": "uuid-del-reporte",
  "timestamp": "2026-07-03T14:30:00Z",
  "id_sensor": "SENSOR-001",
  "ubicacion": { "lat": -33.456, "lon": -70.648 },
  "magnitud": 4.2,
  "profundidad_km": 35,
  "confianza": 0.87
}
```

### Evento: `validacion_positiva`

Publicado por el **Servicio 2: Validación** tras confirmar el sismo con múltiples sensores.

```json
{
  "evento": "validacion_positiva",
  "timestamp": "2026-07-03T14:30:02Z",
  "id_senal": "uuid-de-la-senal-original",
  "sensores_confirmados": ["SENSOR-001", "SENSOR-015", "SENSOR-032"],
  "magnitud_final": 4.3,
  "epicentro": { "lat": -33.458, "lon": -70.650 }
}
```

### Evento: `alerta_emitida`

Publicado por el **Servicio 3: Logística** al completar el registro y la activación de los protocolos de emergencia.

```json
{
  "evento": "alerta_emitida",
  "timestamp": "2026-07-03T14:30:05Z",
  "id_validacion": "uuid-de-la-validacion",
  "nivel_alerta": "AMARILLO",
  "zonas_afectadas": ["Santiago Centro", "Providencia"],
  "costo_emergencia": 150000,
  "estado": "EMITIDA"
}
```

### Niveles de alerta

| Magnitud | Nivel | Costo de emergencia |
|----------|-------|---------------------|
| >= 6.0 | ROJO | $1,500,000 CLP |
| >= 4.0 | AMARILLO | $150,000 CLP |
| < 4.0 | VERDE | $50,000 CLP |

### Flujo completo del evento

```text
Sensor -> POST /api/reportes -> Servicio 1 (Reportes)
  -> Guarda en la BD y publica "senal_recibida" en RabbitMQ
    -> Servicio 2 (Validación) consume, cruza sensores y publica "validacion_positiva"
      -> Servicio 3 (Logística) consume, registra la alerta y publica "alerta_emitida"
        -> Servicio 1 consume "alerta_emitida" y actualiza el estado a EMITIDA
```

---

## CI/CD

| Rama | Entorno | Tag de imagen |
|------|----------|---------------|
| `develop` | QA (`sat-qa`) | `qa-latest` |
| `main` | PROD (`sat-prod`) | `prod-latest` |

Los despliegues serán completamente automáticos mediante GitHub Actions. Se prohíbe el acceso manual a los servidores.

---

## Estructura del Proyecto

```text
SAT/
├── .github/
│   ├── workflows/
│   │   ├── ci-qa.yml          # CI/CD: develop -> QA
│   │   └── ci-prod.yml        # CI/CD: main -> PROD
│   └── CODEOWNERS
├── backend/
│   ├── servicio-reportes/     # Centro de Reportes (REST + Eventos)
│   ├── servicio-validacion/   # Validación de Sensores (Event Driven)
│   └── servicio-logistica/    # Logística de Notificación (Event Driven)
├── frontend/                  # React + Nginx
├── nginx/                     # API Gateway
├── kubernetes/                # Manifiestos K8s
│   ├── namespace.yml
│   ├── configmap.yml
│   ├── secret.yml
│   ├── rabbitmq/
│   ├── postgresql/
│   ├── servicios/
│   ├── frontend/
│   ├── nginx-gateway/
│   ├── ingress.yml
│   ├── backup-cronjob.yml
│   └── logging/
└── scripts/
    └── setup-runner.sh        # Instalación de self-hosted runner
```

