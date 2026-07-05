import { useTheme } from '../theme'
import { useLanguage } from '../i18n'
import BrandMark from '../components/BrandMark'
import SkillCard from '../components/SkillCard'
import { skills } from '../data/skills'

/**
 * Landing screen. Pick a skill to start a session.
 *   onSelectSkill(skill) -> route into the chosen quest
 * The hero box is a placeholder — drop in an <img> or illustration.
 */
export default function Landing({ onSelectSkill, showHero = true }) {
  const { theme: t } = useTheme()
  const { t: tr } = useLanguage()

  return (
    <div style={{ width: '100%', maxWidth: 620 }}>
      <div style={{ marginBottom: 26 }}>
        <BrandMark />
      </div>

      {showHero && (
        <div
          style={{
            width: '100%',
            aspectRatio: '16 / 9',
            borderRadius: 20,
            border: `1px solid ${t.border}`,
            backgroundImage: `repeating-linear-gradient(45deg, ${t.heroA} 0, ${t.heroA} 11px, ${t.heroB} 11px, ${t.heroB} 22px)`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            marginBottom: 28,
          }}
        >
          <span
            style={{
              fontFamily: "ui-monospace, 'SF Mono', Menlo, monospace",
              fontSize: 12.5,
              color: t.heroLabelText,
              background: t.heroLabelBg,
              border: `1px dashed ${t.heroLabelBorder}`,
              borderRadius: 8,
              padding: '7px 13px',
              letterSpacing: '.02em',
            }}
          >
            {tr('landing.hero')}
          </span>
        </div>
      )}

      <h1
        style={{
          fontFamily: "'Bricolage Grotesque', sans-serif",
          fontWeight: 700,
          fontSize: 'clamp(27px, 6vw, 34px)',
          lineHeight: 1.08,
          margin: '0 0 8px',
          letterSpacing: '-.015em',
          textWrap: 'balance',
          color: t.ink,
        }}
      >
        {tr('landing.title')}
      </h1>
      <p style={{ margin: '0 0 26px', fontSize: 15, color: t.soft, maxWidth: '44ch' }}>
        {tr('landing.subtitle')}
      </p>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 13 }}>
        {skills.map((skill) => (
          <SkillCard key={skill.key} skill={skill} onClick={onSelectSkill} />
        ))}
      </div>
    </div>
  )
}
