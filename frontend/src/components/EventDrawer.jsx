// flujo de microservicios 
const TIMELINE_STEPS = [
  { label: 'Centro de Reportes (Microservicio 1)', desc: 'Reporte recibido mediante REST.' },
  { label: 'Validación de Sensores (Microservicio 2)', desc: 'Comparando datos con sensores cercanos.' },
  { label: 'Logística de Notificación (Microservicio 3)', desc: 'Registro oficial del evento y activación de protocolos.' },
  { label: 'Alerta Emitida', desc: '' },
]

// Cuántos pasos están completos según el estado
const DONE_STEPS = { EMITIDA: 4 }
// Qué paso está en progreso (1-indexed, null si ninguno)
const IN_PROGRESS_STEP = { EMITIDA: null }

// Colores del estado en el resumen
const STATUS_TEXT = {
  EMITIDA: 'text-green-700',
}

function formatHora(isoTimestamp) {
  if (!isoTimestamp) return '--'
  return new Date(isoTimestamp).toLocaleTimeString('es-CL')
}

function TimelineStep({ step, index, status }) {
  const stepNum = index + 1
  const done = DONE_STEPS[status] ?? 0
  const inProgress = IN_PROGRESS_STEP[status]

  const isDone = stepNum <= done
  const isInProgress = stepNum === inProgress

  return (
    <div className="relative">
      <div
        className={`absolute -left-[28px] top-1 w-4 h-4 rounded-full border-2 flex items-center justify-center ${
          isDone ? 'border-primary bg-primary' : 'border-outline-variant bg-surface'
        }`}
      >
        {isDone && (
          <span className="material-symbols-outlined text-[10px] text-on-primary">check</span>
        )}
        {isInProgress && (
          <span className="material-symbols-outlined text-[10px] text-primary status-pulse">sync</span>
        )}
      </div>

      <p className={`text-body-md font-bold ${isDone || isInProgress ? 'text-primary' : 'text-secondary'}`}>
        {step.label}
      </p>
      {step.desc && (
        <p className="text-label-sm text-secondary">{step.desc}</p>
      )}
    </div>
  )
}

export default function EventDrawer({ event, onClose }) {
  const isOpen = event !== null

  const summaryFields = event
    ? [
        ['ID', `#${event.id}`],
        ['VALIDACIÓN', event.id_validacion],
        ['MAGNITUD', event.magnitud != null ? `${event.magnitud} Mw` : '--'],
        ['UBICACIÓN', `${event.epicentro_lat}, ${event.epicentro_lon}`],
        ['HORA', formatHora(event.creado_en)],
        ['ESTADO', event.estado],
      ]
    : []

  return (
    <>
      {/* Overlay */}
      <div
        className={`fixed inset-0 bg-black/40 backdrop-blur-sm z-50 transition-opacity ${isOpen ? '' : 'hidden'}`}
        onClick={onClose}
      />

      {/* Panel lateral */}
      <div
        className={`fixed top-0 right-0 h-full w-full max-w-md bg-surface border-l border-outline-variant shadow-2xl z-50 drawer-transition flex flex-col ${
          isOpen ? 'translate-x-0' : 'translate-x-full'
        }`}
      >
        {/* Header */}
        <div className="p-lg border-b border-outline-variant flex items-center justify-between bg-surface-container-low">
          <div>
            <h3 className="text-headline-md text-primary">Detalle del Evento Sísmico</h3>
            <p className="text-label-md text-secondary uppercase mt-1">
              REPORTE #{event?.id ?? '0000'}
            </p>
          </div>
          <button
            className="material-symbols-outlined text-secondary hover:text-primary"
            onClick={onClose}
          >
            close
          </button>
        </div>

        {/* Contenido */}
        {event && (
          <div className="flex-1 overflow-y-auto p-lg space-y-xl">
            {/* Resumen */}
            <div className="grid grid-cols-2 gap-md p-md bg-surface-container rounded-lg">
              {summaryFields.map(([label, value]) => (
                <div key={label}>
                  <p className="text-label-sm text-on-surface-variant">{label}</p>
                  <p
                    className={`text-body-lg font-bold ${
                      label === 'ESTADO' ? (STATUS_TEXT[event.estado] ?? '') : ''
                    }`}
                  >
                    {value}
                  </p>
                </div>
              ))}
            </div>

            {/* Línea de tiempo */}
            <div>
              <h4 className="text-label-md text-secondary border-b border-outline-variant pb-xs mb-md uppercase">
                Línea de Tiempo del Proceso
              </h4>
              <div className="relative pl-8 space-y-xl">
                <div className="absolute left-[11px] top-2 bottom-2 w-[2px] bg-outline-variant" />
                {TIMELINE_STEPS.map((step, i) => (
                  <TimelineStep key={i} step={step} index={i} status={event.estado} />
                ))}
              </div>
            </div>

            {/* Detalles técnicos */}
            <div className="space-y-md pt-lg border-t border-outline-variant">
              <div className="flex justify-between items-center">
                <span className="text-body-md text-secondary">Zonas Afectadas</span>
                <span className="text-label-md text-on-surface bg-outline-variant/20 px-2 py-0.5 rounded">
                  {event.zonas_afectadas?.join(', ') || '--'}
                </span>
              </div>
              {event.costo_emergencia != null && (
                <div className="flex justify-between items-center">
                  <span className="text-body-md text-secondary">Costo de Emergencia</span>
                  <span className="text-body-md font-bold text-green-600">
                    {event.costo_emergencia}
                  </span>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Footer */}
        <div className="p-lg bg-surface-container-low border-t border-outline-variant flex gap-md">
          <button className="flex-1 bg-primary-container text-on-primary py-sm rounded-lg font-bold hover:opacity-90 transition-all">
            Ver Datos Raw
          </button>
          <button
            className="flex-1 bg-surface border border-outline-variant text-on-surface py-sm rounded-lg hover:bg-surface-container transition-colors"
            onClick={onClose}
          >
            Cerrar
          </button>
        </div>
      </div>
    </>
  )
}
