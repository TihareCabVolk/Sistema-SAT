import { useState } from 'react'
import { submitReport } from '../api/seismicApi'

export default function NewReportForm({ onSuccess }) {
  const [sensorId, setSensorId] = useState('')
  const [magnitude, setMagnitude] = useState('')
  const [location, setLocation] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    try {
      await submitReport({ sensorId, magnitude: parseFloat(magnitude), location })
      setSensorId('')
      setMagnitude('')
      setLocation('')
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
        <span className="material-symbols-outlined text-primary-container">add_alert</span>
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

        <div className="space-y-xs">
          <label className="text-label-md text-on-surface-variant block">UBICACIÓN</label>
          <input
            className="w-full bg-surface border border-outline-variant px-md py-sm rounded-lg focus:ring-2 focus:ring-primary-container focus:border-primary-container outline-none text-body-md"
            placeholder="Coordenadas o Región"
            type="text"
            value={location}
            onChange={(e) => setLocation(e.target.value)}
            required
          />
        </div>

        <button
          className="w-full bg-primary-container text-on-primary py-md rounded-lg font-bold hover:opacity-90 transition-all flex justify-center items-center gap-sm mt-lg disabled:opacity-60"
          type="submit"
          disabled={loading}
        >
          <span className={`material-symbols-outlined ${loading ? 'status-pulse' : ''}`}>
            {loading ? 'sync' : 'send'}
          </span>
          {loading ? 'Enviando...' : 'Enviar Reporte'}
        </button>
      </form>
    </section>
  )
}
