const LOGISTICA_URL = '/api/logistica'
const REPORTES_URL = '/api/reportes'

// devuelve todas las alertas sismicas registradas para mostrarlas en el frontend
export async function getAlerts() {
  const res = await fetch(`${LOGISTICA_URL}/alertas`)
  if (!res.ok) throw new Error('Error al obtener alertas')
  return res.json()
}

// envio de un reporte nuevo, y este dispara todo el flujo
export async function submitReport(data) {
  const res = await fetch(`${REPORTES_URL}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error('Error al enviar reporte')
  return res.json()
}
