import { useState } from 'react'
import { useLanguage } from '../i18n'
import BrandLogo from '../components/BrandLogo'
import TextField from '../components/TextField'
import SocialButton from '../components/SocialButton'
import { providers } from '../data/providers'
import { useHover } from '../hooks/useHover'
import { glassDark, roundCorners, hudColors, glassTextShadow } from '../hud'

/**
 * Login screen. All behavior is passed in via props so you can wire it to
 * your backend without touching the markup:
 *   onSubmit({ email, password })  -> POST to your auth endpoint / Cognito
 *   onProvider(providerId)         -> kick off OAuth ('google' | 'facebook')
 *   onForgot(), onSignup()         -> navigation
 */
export default function Login({ onSubmit, onProvider, onForgot, onSignup, showSocial = true }) {
  const { t: tr } = useLanguage()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [busyProvider, setBusyProvider] = useState('')
  const [error, setError] = useState('')
  const [hovered, hoverBind] = useHover()

  const busy = loading || !!busyProvider

  const submit = async (e) => {
    e.preventDefault()
    if (!onSubmit || busy) return
    setError('')
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

  // Gold RPG call to action: dark ink text on the HUD gold, pixel-frame border
  // like the HUD chips. Hover brightens to the XP-stripe highlight gold.
  const primaryBtn = {
    width: '100%',
    boxSizing: 'border-box',
    fontFamily: 'inherit',
    fontSize: 15,
    fontWeight: 700,
    color: hudColors.ink,
    background: hovered ? hudColors.goldLight : hudColors.gold,
    border: '2px solid rgba(255,255,255,0.6)',
    borderRadius: 12,
    padding: 14,
    cursor: 'pointer',
    letterSpacing: '.02em',
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
          ...glassDark,
          ...roundCorners,
          padding: 'clamp(26px, 5.5vw, 40px)',
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 22 }}>
          <BrandLogo fleurSize={40} fontSize={48} letterSpacing={5} gap={12} textShadow={glassTextShadow} />
        </div>

        <h1
          style={{
            fontFamily: "'Jacquard 12', serif",
            fontWeight: 400,
            fontSize: 30,
            margin: '0 0 4px',
            color: '#ffffff',
            textShadow: glassTextShadow,
          }}
        >
          {tr('auth.welcome')}
        </h1>
        <p style={{ margin: '0 0 24px', fontSize: 14, color: 'rgba(255,255,255,0.85)', textShadow: glassTextShadow }}>{tr('auth.welcome.sub')}</p>

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
          autoComplete="current-password"
          gap={22}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          labelColor="#ffffff"
          labelShadow={glassTextShadow}
        />

        {/* Forgot-password flow isn't built yet — hidden so there's no dead button.
            Re-enable this block (and wire the onForgot prop) when password reset lands.
        <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 22 }}>
          <button type="button" onClick={onForgot} style={{ ...linkBtn, fontSize: 12.5 }}>
            Forgot password?
          </button>
        </div>
        */}

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
          {loading ? tr('auth.loggingIn') : tr('auth.login')}
        </button>

        {showSocial && (
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12, margin: '22px 0' }}>
              <span style={{ flex: 1, height: 1, background: 'rgba(255,255,255,0.25)' }} />
              <span style={{ fontSize: 11.5, color: 'rgba(255,255,255,0.75)', letterSpacing: '.04em', textShadow: glassTextShadow }}>{tr('auth.orContinue')}</span>
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
        {tr('auth.newAdventurer')}{' '}
        <button type="button" onClick={onSignup} style={{ ...linkBtn, fontWeight: 700, fontSize: 13.5 }}>
          {tr('auth.createOne')}
        </button>
      </p>
    </form>
  )
}
