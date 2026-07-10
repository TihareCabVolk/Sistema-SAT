import { useState } from 'react'
import { submitReport } from '../api/seismicApi'

export default function NewReportForm({ onSuccess }) {
  const [sensorId, setSensorId] = useState('')
  const [magnitude, setMagnitude] = useState('')
  const [lat, setLat] = useState('')
  const [lon, setLon] = useState('')
  const [profundidadKm, setProfundidadKm] = useState('')
  const [confianza, setConfianza] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    try {
      await submitReport({
        id_sensor: sensorId,
        ubicacion: { lat: parseFloat(lat), lon: parseFloat(lon) },
        magnitud: parseFloat(magnitude),
        profundidad_km: parseInt(profundidadKm, 10),
        confianza: parseFloat(confianza),
        timestamp: new Date().toISOString(),
      })
      setSensorId('')
      setMagnitude('')
      setLat('')
      setLon('')
      setProfundidadKm('')
      setConfianza('')
      onSuccess?.()
    } catch (err) {
      console.error('Error al enviar reporte:', err)
    } finally {
      setLoading(false)
    }
  }

  return (
    <section className="lg:col-span-4 bg-surface-container-lowest border border-outline-variant p-lg rounded-xl flex flex-col gap-lg">
      <div className="flex items-center gap-sm">
        <h2 className="text-headline-md text-on-surface">Nuevo Reporte Sísmico</h2>
      </div>

      <form className="space-y-md" onSubmit={handleSubmit}>
        <div className="space-y-xs">
          <label className="text-label-md text-on-surface-variant block">ID DEL SENSOR</label>
          <input
            className="w-full bg-surface border border-outline-variant px-md py-sm rounded-lg focus:ring-2 focus:ring-primary-container focus:border-primary-container outline-none text-body-md"
            placeholder="Ej. SEN-492-CL"
            type="text"
            value={sensorId}
            onChange={(e) => setSensorId(e.target.value)}
            required
          />
        </div>

        <div className="space-y-xs">
          <label className="text-label-md text-on-surface-variant block">MAGNITUD (Mw)</label>
          <input
            className="w-full bg-surface border border-outline-variant px-md py-sm rounded-lg focus:ring-2 focus:ring-primary-container focus:border-primary-container outline-none text-body-md"
            placeholder="0.0"
            type="number"
            step="0.1"
            min="0"
            max="10"
            value={magnitude}
            onChange={(e) => setMagnitude(e.target.value)}
            required
          />
        </div>

        <div className="grid grid-cols-2 gap-md">
          <div className="space-y-xs">
            <label className="text-label-md text-on-surface-variant block">LATITUD</label>
            <input
              className="w-full bg-surface border border-outline-variant px-md py-sm rounded-lg focus:ring-2 focus:ring-primary-container focus:border-primary-container outline-none text-body-md"
              placeholder="-33.45"
              type="number"
              step="0.0001"
              value={lat}
              onChange={(e) => setLat(e.target.value)}
              required
            />
          </div>
          <div className="space-y-xs">
            <label className="text-label-md text-on-surface-variant block">LONGITUD</label>
            <input
              className="w-full bg-surface border border-outline-variant px-md py-sm rounded-lg focus:ring-2 focus:ring-primary-container focus:border-primary-container outline-none text-body-md"
              placeholder="-70.66"
              type="number"
              step="0.0001"
              value={lon}
              onChange={(e) => setLon(e.target.value)}
              required
            />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-md">
          <div className="space-y-xs">
            <label className="text-label-md text-on-surface-variant block">PROFUNDIDAD (KM)</label>
            <input
              className="w-full bg-surface border border-outline-variant px-md py-sm rounded-lg focus:ring-2 focus:ring-primary-container focus:border-primary-container outline-none text-body-md"
              placeholder="10"
              type="number"
              min="0"
              value={profundidadKm}
              onChange={(e) => setProfundidadKm(e.target.value)}
              required
            />
          </div>
          <div className="space-y-xs">
            <label className="text-label-md text-on-surface-variant block">CONFIANZA</label>
            <input
              className="w-full bg-surface border border-outline-variant px-md py-sm rounded-lg focus:ring-2 focus:ring-primary-container focus:border-primary-container outline-none text-body-md"
              placeholder="0.0 - 1.0"
              type="number"
              step="0.01"
              min="0"
              max="1"
              value={confianza}
              onChange={(e) => setConfianza(e.target.value)}
              required
            />
          </div>
        </div>

        <button
          className="w-full bg-primary-container text-on-primary py-md rounded-lg font-bold hover:opacity-90 transition-all flex justify-center items-center gap-sm mt-lg disabled:opacity-60"
          type="submit"
          disabled={loading}
        >
          {loading ? 'Enviando...' : 'Enviar reporte'}
        </button>
      </form>
    </section>
  )
}
