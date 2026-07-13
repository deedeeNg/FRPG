import { useTheme } from '../theme'
import { useLanguage } from '../i18n'

export default function BrandMark({
  name = 'FRPG',
  tagline,
  showTagline = false,
  // Override text colors (e.g. when placed on a dark "liquid glass" surface).
  nameColor,
  taglineColor,
  textShadow,
}) {
  const { theme: t } = useTheme()
  const { t: tr } = useLanguage()
  const taglineText = tagline ?? tr('brand.tagline')
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 11 }}>
      <span
        style={{
          width: 34,
          height: 34,
          flex: 'none',
          background: t.primary,
          borderRadius: 10,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          boxShadow: `0 6px 14px -6px ${t.gemGlow}`,
        }}
      >
        <span style={{ width: 13, height: 13, background: t.gem, transform: 'rotate(45deg)', borderRadius: 2 }} />
      </span>
      <div style={{ display: 'flex', flexDirection: 'column', lineHeight: 1 }}>
        <span
          style={{
            fontFamily: "'Bricolage Grotesque', sans-serif",
            fontWeight: 800,
            fontSize: 19,
            letterSpacing: '.14em',
            color: nameColor ?? t.ink,
            textShadow,
          }}
        >
          {name}
        </span>
        {showTagline && (
          <span style={{ fontSize: 11.5, color: taglineColor ?? t.mute, letterSpacing: '.02em', marginTop: 4, textShadow }}>{taglineText}</span>
        )}
      </div>
    </div>
  )
}
