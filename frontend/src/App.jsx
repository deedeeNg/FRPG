import { useEffect, useState } from 'react'

function App() {
  const [health, setHealth] = useState('checking...')

  useEffect(() => {
    fetch('/api/health')
      .then((res) => res.json())
      .then((data) => setHealth(data.status))
      .catch(() => setHealth('unreachable'))
  }, [])

  return (
    <main style={{ fontFamily: 'system-ui, sans-serif', padding: '2rem' }}>
      <h1>FRPG ⚔️</h1>
      <p>French learning app, RPG style.</p>
      <p>
        Backend health: <strong>{health}</strong>
      </p>
    </main>
  )
}

export default App
