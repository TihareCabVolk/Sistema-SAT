import { Link, useLocation } from 'react-router-dom'

const NAV_LINKS = [
  { to: '/', label: 'Monitoreo' },
  { to: '/reportes', label: 'Reportes' },
  { to: '/historial', label: 'Historial' },
]

export default function TopNavBar() {
  const { pathname } = useLocation()

  return (
    <nav className="bg-surface border-b border-outline-variant flex justify-between items-center w-full px-margin-mobile md:px-margin-desktop py-md sticky top-0 z-40">
      <div className="flex items-center gap-sm">
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

    </nav>
  )
}
