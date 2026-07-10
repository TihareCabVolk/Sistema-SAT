import { useState, useEffect } from 'react'
import { getAlerts } from '../api/seismicApi'
import NewReportForm from '../components/NewReportForm'
import AlertHistory from '../components/AlertHistory'
import EventDrawer from '../components/EventDrawer'
import Toast from '../components/Toast'

export default function Monitoring() {
  const [alerts, setAlerts] = useState([])
  const [selectedEvent, setSelectedEvent] = useState(null)
  const [toast, setToast] = useState(false)

  useEffect(() => {
    getAlerts().then(setAlerts).catch(() => setAlerts([]))
  }, [])

  const handleReportSuccess = () => {
    setToast(true)
    getAlerts().then(setAlerts).catch(() => {})
  }

  return (
    <>
      <main className="max-w-[1200px] mx-auto px-margin-mobile md:px-margin-desktop py-xxl space-y-xxl pb-20 md:pb-xxl">
        <header className="max-w-2xl">
          <h1 className="text-headline-lg text-primary mb-sm">Consola de Control Central</h1>
          <p className="text-body-lg text-secondary">
            Gestión institucional y monitoreo en tiempo real de la red nacional de sensores sísmicos de alta precisión.
          </p>
        </header>

        <div className="grid grid-cols-1 lg:grid-cols-12 gap-gutter items-start">
          <NewReportForm onSuccess={handleReportSuccess} />
          <AlertHistory alerts={alerts} onSelect={setSelectedEvent} />
        </div>
      </main>

      <EventDrawer event={selectedEvent} onClose={() => setSelectedEvent(null)} />

      <Toast
        message="Reporte recibido correctamente"
        isVisible={toast}
        onHide={() => setToast(false)}
      />
    </>
  )
}
