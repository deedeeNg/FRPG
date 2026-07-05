import { useState, useEffect } from 'react'
import { ThemeProvider, useTheme } from './theme'
import { LanguageProvider, useLanguage } from './i18n'
import Login from './pages/Login'
import SignUp from './pages/SignUp'
import Landing from './pages/Landing'
import Map from './pages/Map'
import Learning from './pages/Learning'
import Settings from './pages/Settings'
import ThemeToggle from './components/ThemeToggle'
import Sidebar from './components/Sidebar'
import ConfirmDialog from './components/ConfirmDialog'
import { login, signup, fetchMe } from './api/auth'
import { getGoogleToken } from './api/google'
import { getFacebookToken } from './api/facebook'

// Demo harness + session gate. Unauthenticated: Login / Sign up. Once a valid JWT
// session exists (verified via GET /api/me): side navbar with Map / Learning /
// Settings. In production, replace the switcher with your router and guard
// routes on the session.
function Screen() {
  const { theme: t } = useTheme()
  const { t: tr } = useLanguage()
  const [session, setSession] = useState(null) // { userId, email } | null
  const [screen, setScreen] = useState('login')
  const [ready, setReady] = useState(false)

  // Sidebar collapse state, persisted in a cookie the shadcn way.
  const [sidebarOpen, setSidebarOpen] = useState(() => {
    const m = typeof document !== 'undefined' && document.cookie.match(/(?:^|; )sidebar_state=([^;]+)/)
    return m ? m[1] === 'true' : true
  })
  const toggleSidebar = () =>
    setSidebarOpen((o) => {
      const next = !o
      document.cookie = `sidebar_state=${next}; path=/; max-age=${60 * 60 * 24 * 7}`
      return next
    })

  // On load, verify any stored token against the backend before trusting it.
  useEffect(() => {
    fetchMe().then((user) => {
      if (user) {
        setSession(user)
        setScreen('home')
      }
      setReady(true)
    })
  }, [])

  // shadcn keyboard shortcut: cmd/ctrl+b toggles the sidebar.
  useEffect(() => {
    const onKey = (e) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'b') {
        e.preventDefault()
        toggleSidebar()
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [])

  // After a successful login/sign-up the token is already stored; verify it the
  // same way a returning session is verified, and enter the app.
  const enterApp = async () => {
    const user = await fetchMe()
    setSession(user)
    setScreen(user ? 'home' : 'login')
  }

  const [confirmLogout, setConfirmLogout] = useState(false)

  const logout = () => {
    localStorage.removeItem('frpg_token')
    setSession(null)
    setScreen('login')
    setConfirmLogout(false)
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

  // Nav depends on the session: logged in uses the side navbar (see Sidebar);
  // logged out uses the login/signup tab pair below.
  const navButtons = [['login', tr('auth.tab.login')], ['signup', tr('auth.tab.signup')]]

  if (ready && session) {
    return (
      <div style={{ display: 'flex', height: '100vh', width: '100%', overflow: 'hidden' }}>
        <Sidebar
          open={sidebarOpen}
          onToggle={toggleSidebar}
          active={screen}
          onSelect={setScreen}
          onLogout={() => setConfirmLogout(true)}
        />
        <div
          style={{
            position: 'relative',
            flex: 1,
            height: '100vh',
            overflowY: 'auto',
            background: t.page,
            fontFamily: "'Public Sans', system-ui, sans-serif",
            color: t.ink,
            transition: 'background .2s, color .2s',
          }}
        >
          <ThemeToggle />
          <div style={{ display: 'flex', justifyContent: 'center', padding: '40px 24px 64px' }}>
            {screen === 'map' ? (
              <Map />
            ) : screen === 'learning' ? (
              <Learning />
            ) : screen === 'settings' ? (
              <Settings />
            ) : (
              <Landing onSelectSkill={(skill) => console.log('selected skill', skill.key)} />
            )}
          </div>
        </div>
        <ConfirmDialog
          open={confirmLogout}
          title={tr('logout.title')}
          description={tr('logout.desc')}
          cancelLabel={tr('common.cancel')}
          actionLabel={tr('logout.action')}
          onCancel={() => setConfirmLogout(false)}
          onConfirm={logout}
        />
      </div>
    )
  }

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
          <p style={{ fontSize: 14, color: t.soft }}>{tr('common.loading')}</p>
        ) : screen === 'signup' ? (
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
        )}
      </div>
    </div>
  )
}

export default function App() {
  return (
    <ThemeProvider>
      <LanguageProvider>
        <Screen />
      </LanguageProvider>
    </ThemeProvider>
  )
}
