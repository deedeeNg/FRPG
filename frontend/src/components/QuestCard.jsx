import { useState } from 'react'
import { useLanguage } from '../i18n'
import { hudColors, pxCorners, liquidGlass, liquidGlassHover, pixelated } from '../hud'

// Keep the white text legible over the frosty, low-tint glass.
const textShadow = '0 2px 0 rgba(43,36,64,0.55)'

/**
 * One quest card in the HUD home grid. A frosty "liquid glass" panel with a
 * solid color chip (pixel skill icon), the skill name over its French label,
 * and a gold "START QUEST" call to action. Clicking routes into its lesson flow.
 *   quest: { key, fr, chipColor, icon, iconFilter }
 *   onClick(quest)
 */
export default function QuestCard({ quest, onClick }) {
  const { t: tr } = useLanguage()
  const [hover, setHover] = useState(false)

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={() => onClick && onClick(quest)}
      onKeyDown={(e) => (e.key === 'Enter' || e.key === ' ') && onClick && onClick(quest)}
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      style={{
        ...liquidGlass,
        ...pxCorners,
        background: hover ? liquidGlassHover : liquidGlass.background,
        padding: '8px 20px',
        display: 'flex',
        flexDirection: 'column',
        gap: 4,
        cursor: 'pointer',
        // Snappy, game-like: no transforms, near-instant bg change.
        transition: 'background .08s',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 11 }}>
        <div
          style={{
            width: 28,
            height: 28,
            flex: 'none',
            background: quest.chipColor,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <img
            src={quest.icon}
            alt=""
            style={{ ...pixelated, width: 15, height: 15, filter: quest.iconFilter }}
          />
        </div>
        <div style={{ lineHeight: 1.05 }}>
          <div style={{ fontWeight: 700, fontSize: 15, color: '#ffffff', textShadow }}>{tr('skill.' + quest.key)}</div>
          <div style={{ fontStyle: 'italic', fontSize: 12, color: 'rgba(255,255,255,0.85)', textShadow }}>{quest.fr}</div>
        </div>
      </div>
      <div style={{ fontSize: 11, color: hudColors.gold, fontWeight: 600, letterSpacing: 1, textShadow }}>
        {tr('quest.start')}
      </div>
    </div>
  )
}
