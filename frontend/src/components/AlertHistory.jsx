// Clases de estado hardcodeadas para que Tailwind las incluya en el bundle
const STATUS_STYLES = {
  Validando: 'bg-yellow-100 text-yellow-800',
  Emitida: 'bg-green-100 text-green-800',
  Recibido: 'bg-blue-100 text-blue-800',
  Error: 'bg-red-100 text-red-800',
}

function exportJSON(alerts) {
  const blob = new Blob([JSON.stringify(alerts, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `alertas-${new Date().toISOString().slice(0, 10)}.json`
  a.click()
  URL.revokeObjectURL(url)
}

export default function AlertHistory({ alerts, onSelect }) {
  return (
    <section className="lg:col-span-8 bg-surface-container-lowest border border-outline-variant rounded-xl overflow-hidden">
      <div className="p-lg border-b border-outline-variant flex justify-between items-center bg-surface-container-low">
        <div className="flex items-center gap-sm">
          <span className="material-symbols-outlined text-primary-container">history</span>
          <h2 className="text-headline-md text-on-surface">Historial de Alertas</h2>
        </div>
        <button
          className="flex items-center gap-xs text-primary text-label-md hover:underline"
          onClick={() => exportJSON(alerts)}
        >
          <span className="material-symbols-outlined text-[18px]">download</span>
          EXPORTAR DATA
        </button>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse">
          <thead className="bg-surface-container-low">
            <tr className="text-label-md text-on-surface-variant border-b border-outline-variant">
              <th className="px-lg py-md">ID</th>
              <th className="px-lg py-md">SENSOR</th>
              <th className="px-lg py-md">MAGNITUD</th>
              <th className="px-lg py-md">UBICACIÓN</th>
              <th className="px-lg py-md">ESTADO</th>
              <th className="px-lg py-md">HORA</th>
              <th className="px-lg py-md">ACCIONES</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-outline-variant">
            {alerts.map((alert) => (
              <tr
                key={alert.id}
                className="hover:bg-surface-container transition-colors cursor-pointer text-body-md"
                onClick={() => onSelect(alert)}
              >
                <td className="px-lg py-md font-bold text-primary">#{alert.id}</td>
                <td className="px-lg py-md">{alert.sensor}</td>
                <td className="px-lg py-md">
                  {alert.magnitude != null ? `${alert.magnitude} Mw` : '--'}
                </td>
                <td className="px-lg py-md">{alert.location}</td>
                <td className="px-lg py-md">
                  <span
                    className={`px-sm py-1 text-[10px] font-bold rounded uppercase tracking-wider ${
                      STATUS_STYLES[alert.status] ?? 'bg-gray-100 text-gray-800'
                    }`}
                  >
                    {alert.status}
                  </span>
                </td>
                <td className="px-lg py-md text-secondary">{alert.time}</td>
                <td className="px-lg py-md">
                  <button
                    className="flex items-center gap-xs text-primary hover:underline text-label-md"
                    onClick={(e) => {
                      e.stopPropagation()
                      onSelect(alert)
                    }}
                  >
                    <span className="material-symbols-outlined text-[18px]">visibility</span>
                    Ver Flujo
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  )
}
