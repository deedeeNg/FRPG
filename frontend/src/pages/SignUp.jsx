import { useState } from 'react'
import { useTheme } from '../theme'
import { useLanguage } from '../i18n'
import BrandMark from '../components/BrandMark'
import TextField from '../components/TextField'
import SocialButton from '../components/SocialButton'
import { providers } from '../data/providers'
import { useHover } from '../hooks/useHover'
import { liquidGlass, roundCorners, hudColors, glassTextShadow } from '../hud'

/**
 * Sign-up screen. Mirrors Login; behavior injected via props:
 *   onSubmit({ email, password })  -> create the account (local email/password)
 *   onProvider(providerId)         -> sign up via OAuth ('google' | 'facebook')
 *   onLogin()                      -> navigate to the login screen
 */
export default function SignUp({ onSubmit, onProvider, onLogin, showSocial = true }) {
  const { theme: t } = useTheme()
  const { t: tr } = useLanguage()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [loading, setLoading] = useState(false)
  const [busyProvider, setBusyProvider] = useState('')
  const [error, setError] = useState('')
  const [hovered, hoverBind] = useHover()

  const busy = loading || !!busyProvider

  const submit = async (e) => {
    e.preventDefault()
    if (!onSubmit || busy) return
    setError('')
    if (password.length < 8) {
      setError(tr('error.passwordShort'))
      return
    }
    if (password !== confirm) {
      setError(tr('error.passwordMismatch'))
      return
    }
    setLoading(true)
    try {
      await onSubmit({ email, password })
    } catch (err) {
      setError(err.message || tr('error.generic'))
    } finally {
      setLoading(false)
    }
  }

  const runProvider = async (id) => {
    if (!onProvider || busy) return
    setError('')
    setBusyProvider(id)
    try {
      await onProvider(id)
    } catch (err) {
      setError(err.message || tr('error.generic'))
    } finally {
      setBusyProvider('')
    }
  }

  const primaryBtn = {
    width: '100%',
    fontFamily: 'inherit',
    fontSize: 15,
    fontWeight: 700,
    color: '#FFFFFF',
    background: hovered ? t.primaryHover : t.primary,
    border: 'none',
    borderRadius: 12,
    padding: 14,
    cursor: 'pointer',
    letterSpacing: '.01em',
    transition: 'background .15s',
  }

  const linkBtn = {
    background: 'none',
    border: 'none',
    padding: 0,
    cursor: 'pointer',
    fontFamily: 'inherit',
    color: hudColors.gold,
    fontWeight: 600,
    textShadow: glassTextShadow,
  }

  return (
    <form onSubmit={submit} style={{ width: '100%', maxWidth: 412 }}>
      <div
        style={{
          ...liquidGlass,
          ...roundCorners,
          padding: 'clamp(26px, 5.5vw, 40px)',
        }}
      >
        <div style={{ marginBottom: 26 }}>
          <BrandMark showTagline nameColor="#ffffff" taglineColor="rgba(255,255,255,0.85)" textShadow={glassTextShadow} />
        </div>

        <h1
          style={{
            fontFamily: "'Bricolage Grotesque', sans-serif",
            fontWeight: 700,
            fontSize: 26,
            margin: '0 0 4px',
            letterSpacing: '-.01em',
            color: '#ffffff',
            textShadow: glassTextShadow,
          }}
        >
          {tr('auth.create')}
        </h1>
        <p style={{ margin: '0 0 24px', fontSize: 14, color: 'rgba(255,255,255,0.85)', textShadow: glassTextShadow }}>{tr('auth.create.sub')}</p>

        <TextField
          label={tr('field.email')}
          type="email"
          name="email"
          placeholder="you@example.com"
          autoComplete="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          labelColor="#ffffff"
          labelShadow={glassTextShadow}
        />
        <TextField
          label={tr('field.password')}
          type="password"
          name="password"
          placeholder="••••••••"
          autoComplete="new-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          labelColor="#ffffff"
          labelShadow={glassTextShadow}
        />
        <TextField
          label={tr('field.confirmPassword')}
          type="password"
          name="confirm"
          placeholder="••••••••"
          autoComplete="new-password"
          gap={22}
          value={confirm}
          onChange={(e) => setConfirm(e.target.value)}
          labelColor="#ffffff"
          labelShadow={glassTextShadow}
        />

        {error && (
          <p role="alert" style={{ margin: '0 0 12px', fontSize: 13, fontWeight: 600, color: '#E5484D' }}>
            {error}
          </p>
        )}

        <button
          type="submit"
          disabled={busy}
          aria-busy={loading}
          style={{ ...primaryBtn, ...(busy ? { opacity: 0.7, cursor: 'not-allowed' } : {}) }}
          {...hoverBind}
        >
          {loading ? tr('auth.creatingAccount') : tr('auth.createAccount')}
        </button>

        {showSocial && (
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12, margin: '22px 0' }}>
              <span style={{ flex: 1, height: 1, background: 'rgba(255,255,255,0.25)' }} />
              <span style={{ fontSize: 11.5, color: 'rgba(255,255,255,0.75)', letterSpacing: '.04em', textShadow: glassTextShadow }}>{tr('auth.orSignup')}</span>
              <span style={{ flex: 1, height: 1, background: 'rgba(255,255,255,0.25)' }} />
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 11 }}>
              {providers.map((p) => (
                <SocialButton
                  key={p.id}
                  label={busyProvider === p.id ? tr('common.connecting') : tr('provider.' + p.id)}
                  mark={p.mark}
                  markSize={p.markSize}
                  disabled={busy}
                  onClick={() => runProvider(p.id)}
                />
              ))}
            </div>
          </div>
        )}
      </div>

      <p style={{ textAlign: 'center', fontSize: 13.5, color: 'rgba(255,255,255,0.9)', textShadow: glassTextShadow, margin: '22px 0 0' }}>
        {tr('auth.haveAccount')}{' '}
        <button type="button" onClick={onLogin} style={{ ...linkBtn, fontWeight: 700, fontSize: 13.5 }}>
          {tr('auth.login')}
        </button>
      </p>
    </form>
  )
}
