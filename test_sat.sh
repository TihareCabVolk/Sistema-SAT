#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# Script de prueba del pipeline SAT
# Simula 3+ sensores reportando un sismo en la misma zona
# para que el servicio de validacion confirme el evento.
# ============================================================

REPORTES_URL="${REPORTES_URL:-http://localhost:4001}"
LOGISTICA_URL="${LOGISTICA_URL:-http://localhost:4003}"

# Ubicacion de prueba (Santiago, Chile)
LAT="-33.45"
LON="-70.67"
MAGNITUD="5.5"
PROFUNDIDAD="60"
CONFIANZA="0.95"

# Sensores que reportan el evento
SENSORES=("SENSOR-001" "SENSOR-002" "SENSOR-003")

echo "===================================================="
echo "  Prueba de Pipeline SAT - Simulacion de sensores"
echo "===================================================="
echo ""
echo "Reportes URL:  $REPORTES_URL"
echo "Logistica URL: $LOGISTICA_URL"
echo "Ubicacion:     lat=$LAT, lon=$LON"
echo "Magnitud:      $MAGNITUD Mw"
echo "Sensores:      ${SENSORES[*]}"
echo ""

# --------------- Enviar reportes ---------------
echo "[1/3] Enviando reportes desde cada sensor..."
for sensor in "${SENSORES[@]}"; do
  TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  echo "  -> $sensor ..."

  RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$REPORTES_URL/api/reportes" \
    -H "Content-Type: application/json" \
    -d "{
      \"id_sensor\": \"$sensor\",
      \"ubicacion\": {\"lat\": $LAT, \"lon\": $LON},
      \"magnitud\": $MAGNITUD,
      \"profundidad_km\": $PROFUNDIDAD,
      \"confianza\": $CONFIANZA,
      \"timestamp\": \"$TIMESTAMP\"
    }")

  HTTP_CODE=$(echo "$RESPONSE" | tail -1)
  BODY=$(echo "$RESPONSE" | sed '$d')

  if [ "$HTTP_CODE" = "202" ]; then
    echo "     [OK] $BODY"
  else
    echo "     [FAIL] HTTP $HTTP_CODE: $BODY"
  fi
done

# --------------- Esperar procesamiento ---------------
echo ""
echo "[2/3] Esperando 5s para que el pipeline procese los eventos..."
sleep 5

# --------------- Consultar alertas generadas ---------------
echo ""
echo "[3/3] Consultando alertas generadas..."
ALERTAS=$(curl -s "$LOGISTICA_URL/api/logistica/alertas" 2>/dev/null || echo "[]")

if command -v jq &>/dev/null; then
  COUNT=$(echo "$ALERTAS" | jq 'length')
  echo ""
  echo "Alertas encontradas: $COUNT"
  echo "===================================================="
  if [ "$COUNT" -gt 0 ]; then
    echo "$ALERTAS" | jq '.[] | {
      id,
      id_validacion,
      magnitud,
      epicentro_lat,
      epicentro_lon,
      nivel_alerta,
      costo_emergencia,
      zonas_afectadas,
      estado,
      creado_en
    }'
  else
    echo "  (sin alertas - revisa que los 3 servicios esten corriendo)"
  fi
else
  echo "$ALERTAS"
  echo ""
  echo "(instala jq para mejor formato: sudo apt install jq)"
fi

echo "===================================================="
echo "Prueba completada."
