#!/bin/bash
# Script de prueba para SAT - Sistema de Alerta Temprana de Sismos
# Simula sensores enviando señales para probar el flujo completo
# 
# Uso: ./scripts/test_sat.sh [URL_base]
# Ej:  ./scripts/test_sat.sh http://localhost:4001
#      ./scripts/test_sat.sh http://qa.grupo2.uta.cl

URL="${1:-http://localhost:4001}"
ENDPOINT="$URL/api/reportes"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}╔════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║   SAT - Prueba de Flujo Completo       ║${NC}"
echo -e "${CYAN}║   Endpoint: $ENDPOINT${NC}"
echo -e "${CYAN}╚════════════════════════════════════════╝${NC}"
echo ""

# ============================================================
# PRUEBA 1: Sensor único (NO debería generar alerta)
# ============================================================
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  PRUEBA 1: Sensor único (sin alerta esperada)${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

SENSOR_UNICO='{
  "id_sensor": "SENSOR-001",
  "ubicacion": { "lat": -33.456, "lon": -70.648 },
  "magnitud": 4.2,
  "profundidad_km": 35,
  "confianza": 0.87,
  "timestamp": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'"
}'

echo -e "Enviando señal de ${YELLOW}SENSOR-001${NC} (magnitud 4.2)..."
RESPONSE=$(curl -s -X POST "$ENDPOINT" \
  -H "Content-Type: application/json" \
  -d "$SENSOR_UNICO")
echo -e "  Respuesta: $RESPONSE"
echo ""

# ============================================================
# PRUEBA 2: 3 sensores cercanos (DEBERÍA generar alerta)
# ============================================================
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}  PRUEBA 2: 3 sensores cercanos (alerta esperada)${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

NOW=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Todos los sensores ubicados en Santiago (~2km entre ellos)
declare -a SENSORS=(
  '{"id_sensor": "SENSOR-A1", "ubicacion": {"lat": -33.456, "lon": -70.648}, "magnitud": 4.2, "profundidad_km": 30, "confianza": 0.92}'
  '{"id_sensor": "SENSOR-B2", "ubicacion": {"lat": -33.460, "lon": -70.652}, "magnitud": 4.3, "profundidad_km": 32, "confianza": 0.88}'
  '{"id_sensor": "SENSOR-C3", "ubicacion": {"lat": -33.452, "lon": -70.644}, "magnitud": 4.4, "profundidad_km": 28, "confianza": 0.95}'
)

for SENSOR in "${SENSORS[@]}"; do
  ID=$(echo "$SENSOR" | grep -o '"id_sensor": "[^"]*"' | cut -d'"' -f4)
  MAG=$(echo "$SENSOR" | grep -o '"magnitud": [^,]*' | cut -d' ' -f2)
  echo -e "Enviando señal de ${GREEN}$ID${NC} (magnitud $MAG)..."
  
  PAYLOAD=$(echo "$SENSOR" | sed 's/.$//')  # quita }
  PAYLOAD="$PAYLOAD, \"timestamp\": \"$NOW\"}"
  
  RESPONSE=$(curl -s -X POST "$ENDPOINT" \
    -H "Content-Type: application/json" \
    -d "$PAYLOAD")
  echo -e "  → $RESPONSE"
  sleep 1
done
echo ""

# ============================================================
# PRUEBA 3: Sismo fuerte (ROJO)
# ============================================================
echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${RED}  PRUEBA 3: Sismo fuerte ≥ 6.0 (alerta ROJO)${NC}"
echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

NOW=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

declare -a SENSORS_FUERTE=(
  '{"id_sensor": "SENSOR-X1", "ubicacion": {"lat": -33.470, "lon": -70.670}, "magnitud": 6.5, "profundidad_km": 15, "confianza": 0.99}'
  '{"id_sensor": "SENSOR-Y2", "ubicacion": {"lat": -33.475, "lon": -70.675}, "magnitud": 6.4, "profundidad_km": 18, "confianza": 0.97}'
  '{"id_sensor": "SENSOR-Z3", "ubicacion": {"lat": -33.465, "lon": -70.665}, "magnitud": 6.6, "profundidad_km": 12, "confianza": 0.98}'
  '{"id_sensor": "SENSOR-W4", "ubicacion": {"lat": -33.472, "lon": -70.672}, "magnitud": 6.3, "profundidad_km": 20, "confianza": 0.96}'
)

for SENSOR in "${SENSORS_FUERTE[@]}"; do
  ID=$(echo "$SENSOR" | grep -o '"id_sensor": "[^"]*"' | cut -d'"' -f4)
  MAG=$(echo "$SENSOR" | grep -o '"magnitud": [^,]*' | cut -d' ' -f2)
  echo -e "Enviando señal de ${RED}$ID${NC} (magnitud $MAG)..."
  
  PAYLOAD=$(echo "$SENSOR" | sed 's/.$//')
  PAYLOAD="$PAYLOAD, \"timestamp\": \"$NOW\"}"
  
  RESPONSE=$(curl -s -X POST "$ENDPOINT" \
    -H "Content-Type: application/json" \
    -d "$PAYLOAD")
  echo -e "  → $RESPONSE"
  sleep 1
done
echo ""

# ============================================================
# VERIFICACIÓN
# ============================================================
echo -e "${CYAN}╔════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║   VERIFICACIÓN RÁPIDA                   ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════╝${NC}"
echo ""
echo -e "Revisa el historial de alertas en:"
echo -e "  ${CYAN}http://localhost:3000${NC} (o la URL del frontend)"
echo ""
echo -e "O consulta la API de logística directamente:"
echo -e "  curl ${CYAN}$URL/api/logistica/alertas${NC} | jq"
echo ""
echo -e "RabbitMQ Management:"
echo -e "  ${CYAN}http://localhost:15672${NC} (guest/guest)"
echo ""
echo -e "${YELLOW}Nota:${NC} Si es primera ejecución, la Prueba 1 puede generar alerta"
echo -e "si hay datos previos en la BD dentro de la ventana de 60s."
