import { useEffect } from 'react'

export default function Toast({ message, isVisible, onHide }) {
  useEffect(() => {
    if (!isVisible) return
    const t = setTimeout(onHide, 3000)
    return () => clearTimeout(t)
  }, [isVisible, onHide])

  if (!isVisible) return null

  return (
    <div className="fixed top-20 right-margin-desktop z-50 bg-surface-container-highest border border-outline-variant p-md rounded-lg shadow-lg flex items-center gap-sm animate-bounce">
      <span className="material-symbols-outlined text-green-600">check_circle</span>
      <span className="text-body-md text-on-surface">{message}</span>
    </div>
  )
}
