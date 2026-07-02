import { useTheme } from '../theme'
import { useHover } from '../hooks/useHover'

export default function SocialButton({ label, mark, markSize = 13, onClick, disabled = false }) {
  const { theme: t } = useTheme()
  const [hovered, hoverBind] = useHover()

  const btn = {
    width: '100%',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 10,
    fontFamily: 'inherit',
    fontSize: 14.5,
    fontWeight: 600,
    color: t.ink,
    background: hovered && !disabled ? t.socialHoverBg : t.socialBg,
    border: `1px solid ${hovered && !disabled ? t.socialHoverBorder : t.socialBorder}`,
    borderRadius: 12,
    padding: 12,
    cursor: disabled ? 'not-allowed' : 'pointer',
    opacity: disabled ? 0.7 : 1,
    transition: 'background .15s, border-color .15s',
  }

  return (
    <button type="button" onClick={onClick} disabled={disabled} style={btn} {...hoverBind}>
      <span
        style={{
          width: 22,
          height: 22,
          flex: 'none',
          borderRadius: 6,
          background: t.mono,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontFamily: "'Bricolage Grotesque', sans-serif",
          fontWeight: 800,
          fontSize: markSize,
          color: t.monoText,
        }}
      >
        {mark}
      </span>
      {label}
    </button>
  )
}
