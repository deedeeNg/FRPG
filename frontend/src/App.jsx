import { useState, useEffect } from 'react'
import { ThemeProvider, useTheme } from './theme'
import { LanguageProvider, useLanguage } from './i18n'
import Login from './pages/Login'
import SignUp from './pages/SignUp'
import Home from './pages/Home'
import HudLayout from './components/HudLayout'
import Background from './components/Background'
import ConfirmDialog from './components/ConfirmDialog'
import { glassTextShadow } from './hud'
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

  const logoutDialog = (
    <ConfirmDialog
      open={confirmLogout}
      title={tr('logout.title')}
      description={tr('logout.desc')}
      cancelLabel={tr('common.cancel')}
      actionLabel={tr('logout.action')}
      onCancel={() => setConfirmLogout(false)}
      onConfirm={logout}
    />
  )

  // Signed in: everything lives on the full-bleed pixel-art HUD. Home renders
  // the quest content; the other nav routes (Map / Learning / Settings) are
  // empty placeholder pages on the same landscape for now.
  if (ready && session) {
    const onLogout = () => setConfirmLogout(true)
    return (
      <>
        {screen === 'home' ? (
          <Home
            activeRoute="home"
            onNavigate={setScreen}
            onLogout={onLogout}
            onSelectQuest={() => setScreen('learning')}
          />
        ) : (
          <HudLayout activeRoute={screen} onNavigate={setScreen} onLogout={onLogout} />
        )}
        {logoutDialog}
      </>
    )
  }

  return (
    <div
      className="authScroll"
      style={{
        position: 'relative',
        height: '100vh',
        width: '100%',
        // Scroll only if a card is taller than the viewport; the bar is hidden
        // (see .authScroll rule) so short screens still read as scrollbar-free.
        overflowY: 'auto',
        overflowX: 'hidden',
        fontFamily: "'Public Sans', system-ui, sans-serif",
        color: t.ink,
        transition: 'color .2s',
      }}
    >
      {/* Hide the scrollbar while keeping the element scrollable. */}
      <style>{`.authScroll{scrollbar-width:none;-ms-overflow-style:none}.authScroll::-webkit-scrollbar{width:0;height:0}`}</style>
      <Background />

      {/* Flex-center inside the scroll container: short content centers, tall
          content stays fully reachable (no top clipping like plain centering). */}
      <div
        style={{
          position: 'relative',
          zIndex: 1,
          minHeight: '100%',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '20px 16px',
        }}
      >
      {ready && (
        <div
          style={{
            position: 'relative',
            zIndex: 1,
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

      <div style={{ position: 'relative', zIndex: 1, width: '100%', display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
        {!ready ? (
          <p style={{ fontSize: 14, color: '#ffffff', textShadow: glassTextShadow }}>{tr('common.loading')}</p>
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
