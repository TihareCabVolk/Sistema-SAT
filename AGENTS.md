# AGENTS.md — Sistema SAT

## Core flow (validación → logística)

```
[Earthquake report] → servicio-reportes (:4001)
  → [RabbitMQ event] → servicio-validacion (:4002)
  → JSON response: { valido, grado, ... }
  → [RabbitMQ event] → servicio-logistica (:4003)
  → genera reporte (fecha, grado, emision)
  → imprime mensaje en terminal
  → actualiza registro a estado "emitida"
```

## Repo layout

- `backend/servicio-validacion/` — validation microservice (Gin + RabbitMQ consumer, PostgreSQL)
- `backend/servicio-logistica/` — logistics microservice (Gin + RabbitMQ consumer/publisher, PostgreSQL)
- `backend/servicio-reportes/` — reports intake microservice (stdlib net/http + RabbitMQ, PostgreSQL)
- `kubernetes/` — full K8s manifests
- `.github/workflows/` — CI/CD: `develop`→QA, `main`→prod

## Current state

**All 3 microservices have full Go implementations** with `go.mod`/`go.sum` and multi-stage Dockerfiles.

| Service | Framework | Port | Consumer | Publisher | DB |
|---|---|---|---|---|---|
| reportes | net/http | 4001 | `alerta_emitida` → update estado | `senal_recibida` | `db-reportes` |
| validacion | Gin | 4002 | `senal_recibida` → validate + Haversine | `validacion_positiva` | `db-validacion` |
| logistica | Gin | 4003 | `validacion_positiva` → create alerta | `alerta_emitida` | `db-logistica` |

Pending:
- `frontend/` — directory exists but no implementation yet

## Getting started (writing code)

```bash
# Build a service
cd backend/servicio-<name> && go build ./...

# Run vet
cd backend/servicio-<name> && go vet ./...

# Build Docker image
docker build -t sat-<name>:latest backend/servicio-<name>
```

## Architecture from infra configs

| Service | Port | DB | Purpose |
|---|---|---|---|
| reportes | 4001 | `db-reportes` | Accepts earthquake reports via REST |
| validacion | 4002 | `db-validacion` | Validates seismic data, produces JSON result |
| logistica | 4003 | `db-logistica` | Post-validation: report generation, status updates |

- **Async bus**: RabbitMQ `amqp://rabbitmq:5672`, exchange `sat.events`
- **Env vars** (from ConfigMap `sat-config`): `DB_*_URL`, `SERVICIO_*_URL`, `RABBITMQ_URL`, `RABBITMQ_EXCHANGE`

## Service design notes

- `servicio-validacion` should respond with JSON containing at minimum:
  - `valido` (bool) — whether it's a valid earthquake
  - `grado` (int/float) — magnitude if valid
- `servicio-logistica` consumes the validated result and:
  - Creates a report with `fecha`, `grado`, `emision`
  - Prints a confirmation message to stdout/stderr
  - Updates the database record status to `"emitida"`
- `servicio-validacion` uses `FOR UPDATE SKIP LOCKED` dentro de una transacción para evitar race conditions entre réplicas concurrentes

## Recent changes (backend/fix branch)

| Change | Archivos |
|--------|----------|
| Race condition fix | `servicio-validacion/repository/senal_repository.go`, `service/validacion_service.go` |
| `SERVER_PORT` en ConfigMap + deployments | `kubernetes/configmap.yml`, `kubernetes/servicios/*/deployment.yml` |
| Eliminar `demo.yml` | `.github/workflows/demo.yml` |
| Retry con backoff + graceful shutdown | `servicio-{reportes,validacion,logistica}/cmd/main.go` |
| autoAck=false + DLQ + idempotencia | `servicio-{reportes,validacion,logistica}/consumer/*.go` |
| Fix ignored errors (Scan, Marshal, Unmarshal) | `servicio-logistica/repository/alerta_repository.go`, `servicio-reportes/handler/reporte_handler.go` |

## CI/CD

- `develop` → Docker images tagged `qa-latest` → deployed to `sat-qa` namespace on KIND
- `main` → Docker images tagged `prod-latest` → deployed to `sat-prod` namespace
- Deploy via: `kubectl set image deployment/<svc> <svc>=<image> -n sat-{qa,prod}`
