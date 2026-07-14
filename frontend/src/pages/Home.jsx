import { useLanguage } from '../i18n'
import HudLayout from '../components/HudLayout'
import QuestCard from '../components/QuestCard'
import AttributesPanel from '../components/AttributesPanel'
import { hudColors, pixelated } from '../hud'
import { useWindowWidth, HUD_BREAKPOINT, CARD_COLLAPSE_BREAKPOINT } from '../hooks/useWindowWidth'
import { attributeScores } from '../data/stats'
import { skills } from '../data/skills'
import knight from '../assets/hud/knight.png'
import iconSpeech from '../assets/hud/icon-speech.png'
import iconNote from '../assets/hud/icon-note.png'
import iconBook from '../assets/hud/icon-book.png'
import iconQuill from '../assets/hud/icon-quill.png'

// The four quest cards, in grid order. Chip colors + icons are from the design;
// French labels reuse the shared skills data. `iconNote` is a white sprite on a
// gold chip, so it's darkened with a brightness filter (per the design).
const fr = (key) => skills.find((s) => s.key === key)?.fr
const QUESTS = [
  { key: 'speaking', chipColor: hudColors.blue, icon: iconSpeech, fr: fr('speaking') },
  { key: 'listening', chipColor: hudColors.gold, icon: iconNote, iconFilter: 'brightness(0.2)', fr: fr('listening') },
  { key: 'reading', chipColor: hudColors.green, icon: iconBook, fr: fr('reading') },
  { key: 'writing', chipColor: hudColors.rose, icon: iconQuill, fr: fr('writing') },
]

/**
 * Home screen — pixel-art RPG HUD ("2a — HUD Aventure"). The whole screen sits
 * on a pixel-art landscape; UI surfaces are dark glass panels.
 *
 * At the 1440×980 reference (and any width ≥ {@link HUD_BREAKPOINT}) it renders
 * the pixel-perfect absolute layout: heading left, knight centered, radar right,
 * quest grid across the bottom. Below the breakpoint — where the heading and
 * radar would collide — the same pieces reflow into a centered stacked column.
 *
 *   activeRoute, onNavigate(route), onLogout — HUD nav
 *   onSelectQuest(quest) — a quest card routes into that skill's lesson flow
 *   showKnight, showXpBar — feature flags (both default true)
 */
export default function Home({
  activeRoute = 'home',
  onNavigate,
  onLogout,
  onSelectQuest,
  showKnight = true,
  showXpBar = true,
}) {
  const { t: tr } = useLanguage()
  const width = useWindowWidth()
  const compact = width < HUD_BREAKPOINT
  // On small screens the quest cards start collapsed (icon + title) and expand
  // on tap; a tighter grid minimum lets them sit two-up on phones.
  const collapseCards = width < CARD_COLLAPSE_BREAKPOINT

  // Hard dark pixel-outline (8-directional) so the white title/subtitle stay
  // legible on any day-cycle phase — including the light day sky where a plain
  // drop shadow washed out. `w` is the outline thickness; the last layer keeps a
  // soft drop shadow for depth.
  const ink = hudColors.ink
  const outline = (w) =>
    `${ink} ${w}px 0 0, ${ink} -${w}px 0 0, ${ink} 0 ${w}px 0, ${ink} 0 -${w}px 0,` +
    `${ink} ${w}px ${w}px 0, ${ink} -${w}px ${w}px 0, ${ink} ${w}px -${w}px 0, ${ink} -${w}px -${w}px 0,` +
    `0 ${w + 2}px 0 rgba(43,36,64,0.35)`

  // Shared content pieces — styled fluidly, then positioned per layout below.
  const heading = (
    <div style={compact ? { textAlign: 'center', maxWidth: 640, margin: '0 auto' } : undefined}>
      <h1
        style={{
          fontFamily: "'Jacquard 12', serif",
          fontSize: compact ? 'clamp(34px, 8vw, 72px)' : 82,
          fontWeight: 400,
          color: '#ffffff',
          margin: 0,
          lineHeight: 1,
          textShadow: outline(3),
        }}
      >
        {tr('landing.title')}
      </h1>
      <p
        style={{
          fontSize: compact ? 'clamp(15px, 3.5vw, 22px)' : 22,
          color: '#ffffff',
          margin: compact ? '14px auto 0' : '18px 0 0',
          textShadow: outline(2),
          maxWidth: 520,
        }}
      >
        {tr('landing.subtitle')}
      </p>
    </div>
  )

  const knightEl = showKnight && (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
      <img
        src={knight}
        alt={tr('stats.title')}
        style={{
          ...pixelated,
          width: compact ? 'clamp(88px, 24vw, 140px)' : 140,
          filter: 'drop-shadow(0 6px 0 rgba(43,36,64,0.3))',
        }}
      />
      <div style={{ width: 130, maxWidth: '70%', height: 18, background: 'rgba(43,36,64,0.25)', borderRadius: '50%', marginTop: -12 }} />
    </div>
  )

  const radar = <AttributesPanel scores={attributeScores} />

  const questGrid = (
    <div
      style={{
        width: '100%',
        display: 'grid',
        // 4-up on wide screens; wraps to 2-up / 1-up as width shrinks.
        gridTemplateColumns: `repeat(auto-fit, minmax(${collapseCards ? 140 : 200}px, 1fr))`,
        gap: 16,
      }}
    >
      {QUESTS.map((quest) => (
        <QuestCard key={quest.key} quest={quest} onClick={onSelectQuest} collapsible={collapseCards} />
      ))}
    </div>
  )

  const hudProps = { activeRoute, onNavigate, onLogout, showXpBar }

  // Narrow: single centered column, stacked in reading order, scrollable.
  if (compact) {
    return (
      <HudLayout {...hudProps}>
        {/* relative + zIndex lifts the static column above the fixed Background
            (z-index 0), which otherwise paints over plain text like the heading. */}
        <div style={{ position: 'relative', zIndex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 28, padding: '8px 20px 28px' }}>
          {heading}
          {knightEl}
          <div style={{ width: '100%', maxWidth: 400 }}>{radar}</div>
          {questGrid}
        </div>
      </HudLayout>
    )
  }

  // Desktop: pixel-perfect absolute layout (the reference look).
  return (
    <HudLayout {...hudProps}>
      <div style={{ position: 'absolute', left: 60, top: 200, width: 620, maxWidth: 'calc(100% - 120px)' }}>{heading}</div>

      {showKnight && (
        <div style={{ position: 'absolute', left: '50%', transform: 'translateX(-50%)', bottom: 110 }}>{knightEl}</div>
      )}

      <div style={{ position: 'absolute', right: 20, top: 128 }}>{radar}</div>

      <div style={{ position: 'absolute', left: 20, right: 20, bottom: 20 }}>{questGrid}</div>
    </HudLayout>
  )
}
