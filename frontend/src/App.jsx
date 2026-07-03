import { useState, useEffect } from 'react'
import { ThemeProvider, useTheme } from './theme'
import Login from './pages/Login'
import SignUp from './pages/SignUp'
import Landing from './pages/Landing'
import Profile from './pages/Profile'
import ThemeToggle from './components/ThemeToggle'
import { login, signup, fetchMe } from './api/auth'
import { getGoogleToken } from './api/google'
import { getFacebookToken } from './api/facebook'

// Demo harness + session gate. Unauthenticated: Login / Sign up. Once a valid JWT
// session exists (verified via GET /api/me): Landing / Profile. In production,
// replace the switcher with your router and guard routes on the session.
function Screen() {
  const { theme: t } = useTheme()
  const [session, setSession] = useState(null) // { userId, email } | null
  const [screen, setScreen] = useState('login')
  const [ready, setReady] = useState(false)

  // On load, verify any stored token against the backend before trusting it.
  useEffect(() => {
    fetchMe().then((user) => {
      if (user) {
        setSession(user)
        setScreen('landing')
      }
      setReady(true)
    })
  }, [])

  // After a successful login/sign-up the token is already stored; verify it the
  // same way a returning session is verified, and enter the app.
  const enterApp = async () => {
    const user = await fetchMe()
    setSession(user)
    setScreen(user ? 'landing' : 'login')
  }

  const logout = () => {
    localStorage.removeItem('frpg_token')
    setSession(null)
    setScreen('login')
  }

  // OAuth flow shared by Login and Sign up (find-or-create on the backend).
  const handleProvider = async (id) => {
    const token = id === 'google' ? await getGoogleToken() : await getFacebookToken()
    const jwt = await login(id, { token })
    localStorage.setItem('frpg_token', jwt)
    await enterApp()
  }

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

  // Nav depends on the session: logged out vs logged in.
  const navButtons = session
    ? [['landing', 'Landing'], ['profile', 'Profile']]
    : [['login', 'Login'], ['signup', 'Sign up']]

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

      {ready && (
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
          {navButtons.map(([key, label]) => (
            <button key={key} type="button" onClick={() => setScreen(key)} style={tab(screen === key)}>
              {label}
            </button>
          ))}
        </div>
      )}

      <div style={{ width: '100%', display: 'flex', flexDirection: 'column', alignItems: 'center', margin: 'auto 0' }}>
        {!ready ? (
          <p style={{ fontSize: 14, color: t.soft }}>Loading…</p>
        ) : !session ? (
          screen === 'signup' ? (
            <SignUp
              onSubmit={async ({ email, password }) => {
                // Creates the local account and logs in; SignUp shows any error
                // (e.g. email already registered) in its banner.
                const jwt = await signup(email, password)
                localStorage.setItem('frpg_token', jwt)
                await enterApp()
              }}
              onProvider={handleProvider}
              onLogin={() => setScreen('login')}
            />
          ) : (
            <Login
              onSubmit={async (creds) => {
                // Throws on bad credentials; Login catches it and shows the error.
                const jwt = await login('local', creds)
                localStorage.setItem('frpg_token', jwt)
                await enterApp()
              }}
              onProvider={handleProvider}
              onForgot={() => console.log('forgot password')}
              onSignup={() => setScreen('signup')}
            />
          )
        ) : screen === 'profile' ? (
          <Profile user={session} onLogout={logout} />
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
