import { Link, useLocation } from 'react-router-dom'

const NAV_ITEMS = [
  { to: '/', icon: 'sensors', label: 'Monitoreo' },
  { to: '/reportes', icon: 'assignment_late', label: 'Reportes' },
  { to: '/historial', icon: 'history', label: 'Historial' },
]

export default function BottomNavBar() {
  const { pathname } = useLocation()

  return (
    <div className="fixed bottom-0 left-0 w-full z-50 flex justify-around items-center px-4 py-sm md:hidden bg-surface border-t border-outline-variant">
      {NAV_ITEMS.map(({ to, icon, label }) => {
        const active = pathname === to
        return (
          <Link
            key={to}
            to={to}
            className={`flex flex-col items-center justify-center transition-transform active:scale-95 ${
              active ? 'text-primary font-bold' : 'text-on-surface-variant'
            }`}
          >
            <span className="material-symbols-outlined">{icon}</span>
            <span className="text-label-md">{label}</span>
          </Link>
        )
      })}
    </div>
  )
}
