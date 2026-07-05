import { useTheme } from '../theme'
import { useLanguage } from '../i18n'
import { useHover } from '../hooks/useHover'

/**
 * Profile screen — placeholder for now (shows the signed-in identity + a way
 * out). Fill in real profile content later.
 */
export default function Profile({ user, onLogout }) {
  const { theme: t } = useTheme()
  const { t: tr } = useLanguage()
  const [hovered, hoverBind] = useHover()

  const email = user?.email || ''
  const initial = (email[0] || '?').toUpperCase()

  return (
    <div style={{ width: '100%', maxWidth: 412 }}>
      <div
        style={{
          background: t.surface,
          border: `1px solid ${t.border}`,
          borderRadius: 24,
          boxShadow: t.cardShadow,
          padding: 'clamp(26px, 5.5vw, 40px)',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          textAlign: 'center',
        }}
      >
        <div
          style={{
            width: 96,
            height: 96,
            borderRadius: '50%',
            background: t.mono,
            color: t.monoText,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontFamily: "'Bricolage Grotesque', sans-serif",
            fontWeight: 800,
            fontSize: 44,
            marginBottom: 18,
          }}
        >
          {initial}
        </div>

        <h1
          style={{
            fontFamily: "'Bricolage Grotesque', sans-serif",
            fontWeight: 700,
            fontSize: 26,
            margin: '0 0 4px',
            letterSpacing: '-.01em',
            color: t.ink,
          }}
        >
          {tr('profile.title')}
        </h1>
        {email && <p style={{ margin: '0 0 24px', fontSize: 14, color: t.soft }}>{email}</p>}

        <button
          type="button"
          onClick={onLogout}
          style={{
            width: '100%',
            fontFamily: 'inherit',
            fontSize: 15,
            fontWeight: 700,
            color: t.ink,
            background: hovered ? t.socialHoverBg : t.socialBg,
            border: `1px solid ${hovered ? t.socialHoverBorder : t.socialBorder}`,
            borderRadius: 12,
            padding: 13,
            cursor: 'pointer',
            transition: 'background .15s, border-color .15s',
          }}
          {...hoverBind}
        >
          {tr('profile.logout')}
        </button>
      </div>
    </div>
  )
}
