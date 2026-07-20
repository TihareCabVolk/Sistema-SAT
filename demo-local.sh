#!/usr/bin/env bash
# ==============================================================================
# demo-local.sh — Demostracion local SAT Grupo 2
# ==============================================================================
# Uso:
#   ./demo-local.sh health       → Verificar health endpoints
#   ./demo-local.sh flujo        → Flujo completo (1 sensor, debe mostrar 1/3)
#   ./demo-local.sh validacion   → 3 sensores → sismo confirmado + alerta
#   ./demo-local.sh niveles      → Probar VERDE, AMARILLO, ROJO
#   ./demo-local.sh caida        → Simular caida de validacion
#   ./demo-local.sh db           → Ver datos en las 3 bases de datos
#   ./demo-local.sh todo         → Ejecutar todas las pruebas
#   ./demo-local.sh help         → Esta ayuda
# ==============================================================================
set -euo pipefail

REPORTES_URL="${REPORTES_URL:-http://localhost:4001}"
LOGISTICA_URL="${LOGISTICA_URL:-http://localhost:4003}"
COLOR_GREEN='\033[0;32m'
COLOR_YELLOW='\033[1;33m'
COLOR_RED='\033[0;31m'
COLOR_BLUE='\033[0;34m'
COLOR_RESET='\033[0m'
BOLD='\033[1m'

log_info()  { echo -e "${COLOR_BLUE}[INFO]${COLOR_RESET}  $*"; }
log_ok()    { echo -e "${COLOR_GREEN}[OK]${COLOR_RESET}    $*"; }
log_warn()  { echo -e "${COLOR_YELLOW}[WARN]${COLOR_RESET}  $*"; }
log_error() { echo -e "${COLOR_RED}[ERROR]${COLOR_RESET} $*"; }
log_title() { echo -e "\n${BOLD}${COLOR_BLUE}$*${COLOR_RESET}"; }

enviar_sismo() {
  local sensor_id="$1"; local lat="$2"; local lon="$3"
  local magnitud="$4"; local profundidad="${5:-30}"; local confianza="${6:-0.90}"

  curl -s -X POST "${REPORTES_URL}/api/reportes" \
    -H "Content-Type: application/json" \
    -d "{
      \"id_sensor\": \"${sensor_id}\",
      \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
      \"ubicacion\": {\"lat\": ${lat}, \"lon\": ${lon}},
      \"magnitud\": ${magnitud},
      \"profundidad_km\": ${profundidad},
      \"confianza\": ${confianza}
    }"
}

verificar_servicio() {
  local url="$1"; local nombre="$2"
  local resp=$(curl -s -o /dev/null -w "%{http_code}" "${url}/health")
  if [ "$resp" = "200" ]; then
    log_ok "${nombre} responde (HTTP ${resp})"
  else
    log_error "${nombre} NO responde (HTTP ${resp})"
    return 1
  fi
}

verificar_rabbitmq() {
  if curl -s -o /dev/null -w "%{http_code}" http://localhost:15672 > /dev/null 2>&1; then
    log_ok "RabbitMQ Management UI accesible en http://localhost:15672"
  else
    log_warn "RabbitMQ Management UI no accesible. El sistema funciona igual."
  fi
}

# ==============================================================================
# PRUEBA: health
# ==============================================================================
test_health() {
  log_title "PRUEBA 1: Health Endpoints"
  verificar_servicio "${REPORTES_URL}" "servicio-reportes"
  verificar_servicio "http://localhost:4002" "servicio-validacion"
  verificar_servicio "${LOGISTICA_URL}" "servicio-logistica"
  verificar_rabbitmq
}

# ==============================================================================
# PRUEBA: flujo
# ==============================================================================
test_flujo() {
  log_title "PRUEBA 2: Flujo con 1 sensor (NO debe generar alerta)"

  log_info "Enviando 1 sensor..."
  local resp=$(enviar_sismo "FLUJO-A" "-33.45" "-70.65" "4.5" "30" "0.95")
  local trace_id=$(echo "$resp" | jq -r '.trace_id // "ERROR"')
  log_info "Trace ID: ${trace_id}"

  sleep 3

  log_info "Logs de validacion (debe decir '1/3'):"
  docker logs servicio-validacion --tail=5 2>&1 | grep -E "guardada|confirmado|idempotencia" || log_warn "No se encontraron logs de validacion. Espera unos segundos mas."
}

# ==============================================================================
# PRUEBA: validacion
# ==============================================================================
test_validacion() {
  log_title "PRUEBA 3: Validacion con 3 sensores (DEBE generar alerta)"

  log_info "Enviando 3 sensores cercanos en ubicacion y tiempo..."

  enviar_sismo "VAL-01" "-33.450" "-70.650" "5.2" "35" "0.92" | jq -r '.trace_id' | while read tid; do log_info "VAL-01 → ${tid}"; done
  sleep 1
  enviar_sismo "VAL-02" "-33.452" "-70.648" "5.1" "32" "0.89" | jq -r '.trace_id' | while read tid; do log_info "VAL-02 → ${tid}"; done
  sleep 1
  enviar_sismo "VAL-03" "-33.455" "-70.655" "5.0" "33" "0.91" | jq -r '.trace_id' | while read tid; do log_info "VAL-03 → ${tid}"; done

  sleep 5

  log_info "=== Logs de validacion ==="
  if docker logs servicio-validacion --tail=5 2>&1 | grep "sismo confirmado"; then
    log_ok "Sismo confirmado por validacion"
  else
    log_warn "No se detecto sismo confirmado. Revisa los logs."
  fi

  log_info "=== Logs de logistica ==="
  if docker logs servicio-logistica --tail=5 2>&1 | grep "alerta emitida"; then
    log_ok "Alerta emitida por logistica"
  else
    log_warn "No se detecto alerta emitida. Revisa los logs."
  fi

  log_info "=== Alertas en DB (ultima) ==="
  curl -s "${LOGISTICA_URL}/api/logistica/alertas" | jq '.[-1] // "sin alertas aun"'
}

# ==============================================================================
# PRUEBA: niveles
# ==============================================================================
test_niveles() {
  log_title "PRUEBA 4: Niveles de alerta (VERDE, AMARILLO, ROJO)"

  local espera=70

  log_info "--- VERDE (magnitud 3.5) ---"
  for i in 01 02 03; do
    enviar_sismo "VERDE-${i}" "-33.45" "-70.65" "3.5" "20" "0.88" > /dev/null
    log_info "Enviado VERDE-${i}"
    sleep 1
  done
  sleep 5
  log_info "Ultima alerta generada:"
  curl -s "${LOGISTICA_URL}/api/logistica/alertas" | jq '.[-1] | {nivel: .nivel_alerta, costo: .costo_emergencia} // "esperando..."'

  log_warn "Esperando ${espera}s para nueva ventana de validacion..."
  sleep ${espera}

  log_info "--- AMARILLO (magnitud 4.5) ---"
  for i in 01 02 03; do
    enviar_sismo "AMARILLO-${i}" "-33.45" "-70.65" "4.5" "25" "0.90" > /dev/null
    log_info "Enviado AMARILLO-${i}"
    sleep 1
  done
  sleep 5
  log_info "Ultima alerta generada:"
  curl -s "${LOGISTICA_URL}/api/logistica/alertas" | jq '.[-1] | {nivel: .nivel_alerta, costo: .costo_emergencia} // "esperando..."'

  log_warn "Esperando ${espera}s para nueva ventana de validacion..."
  sleep ${espera}

  log_info "--- ROJO (magnitud 7.1) ---"
  for i in 01 02 03; do
    enviar_sismo "ROJO-${i}" "-33.45" "-70.65" "7.1" "15" "0.95" > /dev/null
    log_info "Enviado ROJO-${i}"
    sleep 1
  done
  sleep 5
  log_info "Ultima alerta generada:"
  curl -s "${LOGISTICA_URL}/api/logistica/alertas" | jq '.[-1] | {nivel: .nivel_alerta, costo: .costo_emergencia} // "esperando..."'
}

# ==============================================================================
# PRUEBA: caida
# ==============================================================================
test_caida() {
  log_title "PRUEBA 5: Simular caida de servicio-validacion"

  log_warn "Deteniendo servicio-validacion..."
  docker stop servicio-validacion
  sleep 2

  log_info "Enviando 3 sensores (quedaran encolados en RabbitMQ)..."
  for i in 01 02 03; do
    enviar_sismo "CAIDA-${i}" "-33.45" "-70.65" "4.5" "25" "0.90" > /dev/null
    log_info "Enviado CAIDA-${i}"
    sleep 1
  done

  log_info "Verifica http://localhost:15672 → cola_señales_recibidas → 3 mensajes encolados"
  log_info "Presiona ENTER para levantar validacion de nuevo..."
  read -r

  docker start servicio-validacion
  log_info "Validacion reiniciado. Esperando procesamiento..."
  sleep 6

  log_info "=== Logs de validacion (debe procesar los 3 acumulados) ==="
  docker logs servicio-validacion --tail=10 2>&1 | grep -E "sismo confirmado|guardada|idempotencia"
}

# ==============================================================================
# PRUEBA: db
# ==============================================================================
test_db() {
  log_title "PRUEBA 6: Datos en las bases de datos"

  log_info "=== DB Reportes (lecturas_sensores) ==="
  docker exec db-reportes psql -U reportes -d reportes \
    -c "SELECT trace_id, id_sensor, magnitud, estado FROM lecturas_sensores ORDER BY timestamp DESC LIMIT 5;" 2>/dev/null || log_warn "No se pudo consultar db-reportes"

  log_info "=== DB Validacion (senales) ==="
  docker exec db-validacion psql -U validacion -d validacion \
    -c "SELECT id, id_sensor, magnitud, validada FROM senales ORDER BY timestamp DESC LIMIT 5;" 2>/dev/null || log_warn "No se pudo consultar db-validacion"

  log_info "=== DB Logistica (alertas) ==="
  docker exec db-logistica psql -U logistica -d logistica \
    -c "SELECT id, nivel_alerta, magnitud, costo_emergencia, estado FROM alertas ORDER BY creado_en DESC LIMIT 5;" 2>/dev/null || log_warn "No se pudo consultar db-logistica"
}

# ==============================================================================
# PRUEBA: todo
# ==============================================================================
test_todo() {
  test_health
  test_flujo
  sleep 2
  test_validacion
  test_db

  echo ""
  log_ok "Pruebas basicas completadas."
  log_info "Para probar niveles de alerta:  ./demo-local.sh niveles"
  log_info "Para simular caida:            ./demo-local.sh caida"
}

# ==============================================================================
# help
# ==============================================================================
show_help() {
  echo "Uso: ./demo-local.sh <comando>"
  echo ""
  echo "Comandos:"
  echo "  health        Verificar health endpoints de los 3 servicios"
  echo "  flujo         Enviar 1 sensor (NO genera alerta, muestra '1/3')"
  echo "  validacion    Enviar 3 sensores cercanos (DEBE generar alerta)"
  echo "  niveles       Probar VERDE (3.5), AMARILLO (4.5), ROJO (7.1)"
  echo "  caida         Simular caida de validacion y recuperacion"
  echo "  db            Mostrar datos en las 3 bases de datos"
  echo "  todo          Ejecutar health + flujo + validacion + db"
  echo "  help          Esta ayuda"
  echo ""
  echo "Variables de entorno:"
  echo "  REPORTES_URL   URL del servicio reportes (default: http://localhost:4001)"
  echo "  LOGISTICA_URL  URL del servicio logistica (default: http://localhost:4003)"
}

# ==============================================================================
# MAIN
# ==============================================================================
case "${1:-help}" in
  health)     test_health ;;
  flujo)      test_flujo ;;
  validacion) test_validacion ;;
  niveles)    test_niveles ;;
  caida)      test_caida ;;
  db)         test_db ;;
  todo)       test_todo ;;
  help|*)     show_help ;;
esac
