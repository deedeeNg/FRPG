import { useState } from 'react'
import { ThemeProvider, useTheme } from './theme'
import Login from './pages/Login'
import Landing from './pages/Landing'
import ThemeToggle from './components/ThemeToggle'
import { login } from './api/auth'

// Demo harness: a page shell + a Login/Landing switcher so you can see both
// screens. In production, replace the switcher with your router (react-router,
// TanStack Router, etc.) and render <Login /> / <Landing /> per route.
function Screen() {
  const { theme: t } = useTheme()
  const [screen, setScreen] = useState('login')

  const tab = (active) => ({
    fontFamily: 'inherit',
    fontSize: 13.5,
    fontWeight: 700,
    border: 'none',
    borderRadius: 999,
    padding: '8px 20px',
    cursor: 'pointer',
    letterSpacing: '.01em',
    transition: 'background .15s, color .15s',
    background: active ? t.tabActiveBg : 'transparent',
    color: active ? t.tabActiveText : t.tabIdle,
    boxShadow: active ? `0 1px 3px ${t.tabShadow}` : 'none',
  })

  return (
    <div
      style={{
        position: 'relative',
        minHeight: '100vh',
        width: '100%',
        background: t.page,
        fontFamily: "'Public Sans', system-ui, sans-serif",
        color: t.ink,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        padding: '20px 16px 48px',
        transition: 'background .2s, color .2s',
      }}
    >
      <ThemeToggle />

      <div
        style={{
          display: 'inline-flex',
          gap: 4,
          padding: 4,
          background: t.tabWrap,
          border: `1px solid ${t.tabWrapBorder}`,
          borderRadius: 999,
          marginBottom: 28,
        }}
      >
        <button type="button" onClick={() => setScreen('login')} style={tab(screen === 'login')}>
          Login
        </button>
        <button type="button" onClick={() => setScreen('landing')} style={tab(screen === 'landing')}>
          Landing
        </button>
      </div>

      <div style={{ width: '100%', display: 'flex', flexDirection: 'column', alignItems: 'center', margin: 'auto 0' }}>
        {screen === 'login' ? (
          <Login
            onSubmit={async (creds) => {
              // Throws on bad credentials; Login catches it and shows the error.
              const token = await login('local', creds)
              localStorage.setItem('frpg_token', token)
              setScreen('landing')
            }}
            onProvider={(id) => console.log('oauth provider', id)} // TODO: needs Google/FB SDK to get a token first
            onForgot={() => console.log('forgot password')}
            onSignup={() => setScreen('login')}
          />
        ) : (
          <Landing onSelectSkill={(skill) => console.log('selected skill', skill.key)} />
        )}
      </div>
    </div>
  )
}

export default function App() {
  return (
    <ThemeProvider>
      <Screen />
    </ThemeProvider>
  )
}
