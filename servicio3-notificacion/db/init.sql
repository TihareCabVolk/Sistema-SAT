-- Historial oficial de alertas (BD exclusiva del Servicio 3)
CREATE TABLE IF NOT EXISTS alertas (
    id                   UUID PRIMARY KEY,
    reporte_id           UUID NOT NULL,
    -- UNIQUE = idempotencia: si RabbitMQ re-entrega el mismo evento,
    -- la segunda inserción falla y el servicio lo descarta con ACK.
    event_id_origen      UUID NOT NULL UNIQUE,
    magnitud             NUMERIC(4,2) NOT NULL,
    nivel                VARCHAR(10) NOT NULL CHECK (nivel IN ('AMARILLA','NARANJA','ROJA')),
    costo_emergencia_clp BIGINT NOT NULL,
    estado               VARCHAR(20) NOT NULL DEFAULT 'EMITIDA',
    creada_en            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_alertas_reporte ON alertas (reporte_id);
CREATE INDEX IF NOT EXISTS idx_alertas_creada ON alertas (creada_en DESC);
