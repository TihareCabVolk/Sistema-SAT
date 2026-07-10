const LOGISTICA_URL = import.meta.env.VITE_LOGISTICA_URL ?? 'http://localhost:4003'
const REPORTES_URL = import.meta.env.VITE_REPORTES_URL ?? 'http://localhost:4001'

// Datos de referencia usados durante el desarrollo inicial del frontend, antes
// de que los servicios backend estuvieran disponibles. Ya no se usan.
const MOCK_ALERTS = [
  {
    id: '6842',
    sensor: 'VLP-01-A',
    magnitude: 4.2,
    location: 'Valparaíso, CL',
    status: 'Validando',
    time: '14:22:10',
    signature: 'SHA-256 VERIFIED',
    latency: '42ms',
  },
  {
    id: '6841',
    sensor: 'ANT-09-C',
    magnitude: 6.8,
    location: 'Antofagasta, CL',
    status: 'Emitida',
    time: '12:05:45',
    signature: 'SHA-256 VERIFIED',
    latency: '38ms',
  },
  {
    id: '6840',
    sensor: 'BIO-02-B',
    magnitude: 3.5,
    location: 'Concepción, CL',
    status: 'Recibido',
    time: '09:12:01',
    signature: 'SHA-256 VERIFIED',
    latency: '55ms',
  },
  {
    id: '6839',
    sensor: 'SCL-15-X',
    magnitude: null,
    location: 'Santiago, CL',
    status: 'Error',
    time: '08:55:30',
    signature: null,
    latency: null,
  },
]
void MOCK_ALERTS

export async function getAlerts() {
  const res = await fetch(`${LOGISTICA_URL}/api/alertas`)
  if (!res.ok) throw new Error('Error al obtener alertas')
  return res.json()
}

export async function submitReport(data) {
  const res = await fetch(`${REPORTES_URL}/api/reportes`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error('Error al enviar reporte')
  return res.json()
}
