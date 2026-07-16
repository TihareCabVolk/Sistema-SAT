# SAT — Sistema de Alerta Temprana de Sismos

## Resumen General

---

## Lógica General

El sistema simula una red de sensores sísmicos. Un sensor envía una señal vía REST, el sistema la valida cruzando datos con otros sensores cercanos, y si se confirma un sismo real, emite una alerta con nivel de severidad y costo de emergencia asociado.

**Flujo completo:**
```
Sensor → POST /api/reportes → servicio-reportes (4001)
  → Guarda en DB → Publica "señal_recibida" en RabbitMQ
    → servicio-validacion (4002) consume
      → Guarda en DB → Busca sensores cercanos (100km, 60s)
      → Si ≥ 3 sensores confirman → Publica "validacion_positiva"
        → servicio-logistica (4003) consume
          → Calcula nivel (ROJO/AMARILLO/VERDE) y costo
          → Guarda alerta en DB → Publica "alerta_emitida"
```

---

## Arquitectura Event-Driven

| Componente | Protocolo | Rol |
|-----------|-----------|-----|
| **servicio-reportes** | REST + Eventos | Recibe señales, guarda y publica a la cola |
| **servicio-validacion** | Event Driven (solo cola) | Consume, cruza datos geográficos, decide si valida |
| **servicio-logistica** | Event Driven + REST (consulta) | Consume, genera alerta final |

**RabbitMQ**: Exchange `sat.events` (topic), 3 routing keys:
- `señal_recibida` → Reportes → Validación
- `validacion_positiva` → Validación → Logística
- `alerta_emitida` → Logística → (futuro: Reportes actualiza estado)

**Consistencia de eventos**: ACK manual + `ON CONFLICT DO NOTHING` → garantía **effectively-once**.

---

## Bases de Datos (3 PostgreSQL independientes)

| Servicio | BD | Tabla | Propósito |
|----------|-----|-------|-----------|
| Reportes | `db-reportes` | `lecturas_sensores` | Almacena señales crudas entrantes |
| Validación | `db-validacion` | `senales` | Almacena señales para cruce geográfico |
| Logística | `db-logistica` | `alertas` | Almacena alertas finales emitidas |

Cada microservicio solo accede a su propia BD — comunicación entre servicios es **exclusivamente por RabbitMQ**.

---

## Kubernetes

| Componente | Tipo | Réplicas | Puerto |
|-----------|------|----------|--------|
| **API Gateway** (Nginx) | Deployment | 2 | 80 |
| **Frontend** (React) | Deployment | 2 | 80 |
| **servicio-reportes** | Deployment | 2 | 4001 |
| **servicio-validacion** | Deployment | 2 | 4002 |
| **servicio-logistica** | Deployment | 2 | 4003 |
| **db-reportes** | StatefulSet (PVC 1Gi) | 1 | 5432 |
| **db-validacion** | StatefulSet (PVC 1Gi) | 1 | 5432 |
| **db-logistica** | StatefulSet (PVC 1Gi) | 1 | 5432 |
| **RabbitMQ** | StatefulSet (PVC 1Gi) | 1 | 5672/15672 |
| **Backup** | CronJob | — | Cada 10min (`pg_dump`) |
| **ELK** (solo QA) | Elasticsearch + Fluent Bit + Kibana | — | 9200/5601 |

**2 namespaces**: `sat-qa` (qa.grupo2.uta.cl) y `sat-prod` (prod.grupo2.uta.cl).

---

## CI/CD

**GitHub Actions** con self-hosted runner en las VMs del departamento:
- `git push develop` → build + deploy a QA
- `git push main` → build + deploy a PROD
- Zero despliegues manuales

---

## Optimización de Imágenes

**Multi-stage build** con Alpine:
- Build: `golang:1.22-alpine`
- Runtime: `alpine:3.19`
- Tamaño final: **~15 MB** por microservicio (vs ~385 MB con imagen Go completa)
- **96% menos espacio**, pulls más rápidos en K8s, menor superficie de ataque
