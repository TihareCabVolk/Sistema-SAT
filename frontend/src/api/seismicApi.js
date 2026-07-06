// Capa de API — reemplazar los mocks con fetch reales cuando el backend esté listo.
// Variables de entorno necesarias: VITE_API_URL (ej. http://localhost:3000)
//
// Endpoints esperados (acordar con el equipo de backend):
//   GET  /api/alerts       -> lista de alertas sísmicas
//   POST /api/reports      -> enviar nuevo reporte sísmico
//   GET  /api/alerts/:id   -> detalle de una alerta

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000'

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

export async function getAlerts() {
  // TODO: descomentar cuando el backend esté disponible
  // const res = await fetch(`${BASE_URL}/api/alerts`)
  // if (!res.ok) throw new Error('Error al obtener alertas')
  // return res.json()
  void BASE_URL
  return Promise.resolve([...MOCK_ALERTS])
}

export async function submitReport(data) {
  // TODO: descomentar cuando el backend esté disponible
  // const res = await fetch(`${BASE_URL}/api/reports`, {
  //   method: 'POST',
  //   headers: { 'Content-Type': 'application/json' },
  //   body: JSON.stringify(data),
  // })
  // if (!res.ok) throw new Error('Error al enviar reporte')
  // return res.json()
  console.info('[MOCK] Enviando reporte al backend:', data)
  return Promise.resolve({ success: true, id: String(Math.floor(Math.random() * 9000) + 1000) })
}
