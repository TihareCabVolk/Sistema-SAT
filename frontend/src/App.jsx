import { BrowserRouter, Routes, Route } from 'react-router-dom'
import TopNavBar from './components/TopNavBar'
import BottomNavBar from './components/BottomNavBar'
import Monitoring from './pages/Monitoring'
import Reports from './pages/Reports'
import History from './pages/History'
import './App.css'

export default function App() {
  return (
    <BrowserRouter>
      <div className="bg-surface text-on-surface min-h-screen">
        <TopNavBar />
        <Routes>
          <Route path="/" element={<Monitoring />} />
          <Route path="/reportes" element={<Reports />} />
          <Route path="/historial" element={<History />} />
        </Routes>
        <BottomNavBar />
      </div>
    </BrowserRouter>
  )
}
