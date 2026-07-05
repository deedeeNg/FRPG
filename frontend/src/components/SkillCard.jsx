import { useTheme } from '../theme'
import { useLanguage } from '../i18n'
import { useHover } from '../hooks/useHover'

const SHAPES = {
  circle: { borderRadius: '50%' },
  diamond: { transform: 'rotate(45deg)', borderRadius: 3 },
  square: { borderRadius: 6 },
  triangle: { clipPath: 'polygon(50% 0, 100% 100%, 0 100%)' },
}

export default function SkillCard({ skill, onClick }) {
  const { theme: t } = useTheme()
  const { t: tr } = useLanguage()
  const [hovered, hoverBind] = useHover()

  const btn = {
    display: 'flex',
    alignItems: 'center',
    gap: 14,
    textAlign: 'left',
    width: '100%',
    background: t.surface,
    border: `1px solid ${hovered ? t.skillHoverBorder : t.border}`,
    borderRadius: 16,
    padding: '18px 16px',
    cursor: 'pointer',
    fontFamily: 'inherit',
    color: t.ink,
    transform: hovered ? 'translateY(-2px)' : 'none',
    boxShadow: hovered ? `0 14px 26px -18px ${t.skillShadow}` : 'none',
    transition: 'transform .12s, box-shadow .15s, border-color .15s',
  }

  return (
    <button type="button" onClick={() => onClick && onClick(skill)} style={btn} {...hoverBind}>
      <span
        style={{
          width: 40,
          height: 40,
          flex: 'none',
          borderRadius: 11,
          background: t.tokenWrap,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <span style={{ width: 22, height: 22, display: 'block', background: skill.color, ...(SHAPES[skill.shape] || SHAPES.circle) }} />
      </span>
      <span style={{ display: 'flex', flexDirection: 'column', lineHeight: 1.15 }}>
        <span style={{ fontSize: 15.5, fontWeight: 700 }}>{tr('skill.' + skill.key)}</span>
        <span style={{ fontSize: 12.5, color: t.mute, fontStyle: 'italic', marginTop: 2 }}>{skill.fr}</span>
      </span>
    </button>
  )
}
