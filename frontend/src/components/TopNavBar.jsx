import { Link, useLocation } from 'react-router-dom'

const NAV_LINKS = [
  { to: '/', label: 'Monitoreo' },
  { to: '/reportes', label: 'Reportes' },
  { to: '/historial', label: 'Historial' },
]

const STATUS_SERVICES = ['API Gateway', 'RabbitMQ', 'Backend']

export default function TopNavBar() {
  const { pathname } = useLocation()

  return (
    <nav className="bg-surface border-b border-outline-variant flex justify-between items-center w-full px-margin-mobile md:px-margin-desktop py-md sticky top-0 z-40">
      <div className="flex items-center gap-sm">
        <span className="material-symbols-outlined text-primary text-3xl">shield_with_heart</span>
        <span className="text-headline-md text-primary font-bold tracking-tight">SEISMIC SHIELD</span>
      </div>

      <div className="hidden md:flex items-center gap-xl">
        {NAV_LINKS.map(({ to, label }) => (
          <Link
            key={to}
            to={to}
            className={`py-xs text-body-lg transition-colors ${
              pathname === to
                ? 'text-primary border-b-2 border-primary'
                : 'text-secondary hover:text-primary'
            }`}
          >
            {label}
          </Link>
        ))}
      </div>

      <div className="flex items-center gap-md">
        <div className="flex items-center gap-xs px-sm py-xs bg-secondary-container rounded-full">
          <span className="w-2 h-2 rounded-full bg-green-600 status-pulse" />
          <span className="text-label-md text-on-secondary-container">Connected</span>
        </div>

        <div className="hidden md:flex items-center gap-xs">
          {STATUS_SERVICES.map((label) => (
            <div key={label} className="flex items-center gap-xs px-sm py-xs bg-green-100 rounded-full">
              <span className="w-2 h-2 rounded-full bg-green-600" />
              <span className="text-label-sm text-green-800">{label}</span>
            </div>
          ))}
        </div>

        <div className="flex gap-sm">
          <button className="material-symbols-outlined text-on-surface-variant hover:text-primary transition-colors cursor-pointer">
            notifications
          </button>
          <button className="material-symbols-outlined text-on-surface-variant hover:text-primary transition-colors cursor-pointer">
            settings
          </button>
        </div>
      </div>
    </nav>
  )
}
