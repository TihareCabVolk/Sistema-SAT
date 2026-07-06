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

- `backend/servicio-validacion/` — validation microservice (mostly empty)
- `backend/servicio-logistica/` — logistics microservice (`cmd/main.go` only)
- `backend/servicio-reportes/` — reports intake microservice (empty)
- `kubernetes/` — full K8s manifests
- `.github/workflows/` — CI/CD: `develop`→QA, `main`→prod

## Current state

**No Go code has been written yet** — most directories contain only a `Dockerfile`.
- No `go.mod`/`go.sum` — run `go mod init` before writing code.
- `.env` is empty; populate with `RABBITMQ_URL`, `SERVICIO_*_URL` as needed.
- The Go backend `dockerfile` at root references Go 1.26; individual service Dockerfiles use Go 1.22.

## Getting started (writing code)

```bash
# Initialize a Go service (do this first)
cd backend/servicio-<name> && go mod init <module-path>

# Build
cd backend/servicio-<name> && go build -o /dev/null ./...

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

## CI/CD

- `develop` → Docker images tagged `qa-latest` → deployed to `sat-qa` namespace on KIND
- `main` → Docker images tagged `prod-latest` → deployed to `sat-prod` namespace
- Deploy via: `kubectl set image deployment/<svc> <svc>=<image> -n sat-{qa,prod}`
