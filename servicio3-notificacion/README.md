# Servicio 3: Logística de Notificación (SAT - Sismos)

Servicio **100% Event-Driven**. No expone API REST pública (solo `/health` para Kubernetes).

## Responsabilidad

1. Consume el evento `sismo.validado` (publicado por Servicio 2).
2. Registra la alerta en el **historial oficial** (su propia BD PostgreSQL).
3. Calcula/activa los protocolos de costos de emergencia.
4. Publica el evento `alerta.emitida` para que el Servicio 1 (Centro de Reportes)
   cambie el estado de la alerta a **"Emitida"**.

## Contrato de Datos (JSON)

### Evento que CONSUME — `sismo.validado`
Cola: `q.servicio3.sismo-validado` (binding key: `sismo.validado`, exchange: `sat.eventos`, tipo topic)

```json
{
  "event_id": "uuid-v4",
  "event_type": "sismo.validado",
  "timestamp": "2026-07-03T14:32:10Z",
  "payload": {
    "reporte_id": "uuid-del-reporte-original",
    "sensor_id": "SENSOR-ARICA-01",
    "magnitud": 6.8,
    "latitud": -18.4783,
    "longitud": -70.3126,
    "profundidad_km": 35.2,
    "sensores_confirmantes": 3
  }
}
```

### Evento que PUBLICA — `alerta.emitida`
Routing key: `alerta.emitida`, exchange: `sat.eventos`

```json
{
  "event_id": "uuid-v4",
  "event_type": "alerta.emitida",
  "timestamp": "2026-07-03T14:32:11Z",
  "payload": {
    "reporte_id": "uuid-del-reporte-original",
    "alerta_id": "uuid-generado-por-servicio3",
    "nivel": "ROJA",
    "costo_emergencia_clp": 15000000,
    "canales_notificados": ["ciudadania", "equipos_emergencia", "autoridades"],
    "estado": "EMITIDA"
  }
}
```

## Variables de entorno

| Variable | Ejemplo | Descripción |
|---|---|---|
| `RABBITMQ_URL` | `amqp://user:pass@rabbitmq:5672/` | Conexión al broker |
| `DATABASE_URL` | `postgres://s3:pass@db-servicio3:5432/notificaciones?sslmode=disable` | BD propia del servicio |
| `HTTP_PORT` | `8083` | Puerto del endpoint /health |

## Ejecución local

```bash
docker compose up -d   # levanta rabbitmq + postgres locales
go run ./cmd
```

## Comandos operativos útiles

```bash
# Ver logs del servicio en el clúster
kubectl logs -l app=servicio3-notificacion -n grupo2-qa -f

# Verificar la cola en RabbitMQ
kubectl exec -it deploy/rabbitmq -n grupo2-qa -- rabbitmqctl list_queues name messages

# Ver historial oficial en la BD
kubectl exec -it deploy/db-servicio3 -n grupo2-qa -- psql -U s3 -d notificaciones -c "SELECT * FROM alertas ORDER BY creada_en DESC LIMIT 10;"
```
