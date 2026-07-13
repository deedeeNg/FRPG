import { useLanguage } from '../i18n'
import StatHexagon from './StatHexagon'
import { hudColors, roundCorners, liquidGlass } from '../hud'

// Grid-ring alpha steps from the design: inner → outer.
const ringAlpha = [0.2, 0.25, 0.35]
const gridStroke = (r) => `rgba(255,255,255,${ringAlpha[r] ?? 0.25})`

/**
 * Attributes radar panel (right side of the HUD home). Dark-glass panel with a
 * gold "ATTRIBUTES" header over the reused {@link StatHexagon}, re-skinned to
 * the pixel-art HUD look (gold polygon on dark glass, white Pixelify labels).
 *   scores: { [attrKey]: 0..MAX_SCORE } — driven from user skill data.
 */
export default function AttributesPanel({ scores }) {
  const { t: tr } = useLanguage()

  return (
    <div
      style={{
        ...liquidGlass,
        ...roundCorners,
        width: 'min(400px, 100%)',
        boxSizing: 'border-box',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        padding: '22px 20px 26px',
      }}
    >
      <div
        style={{
          alignSelf: 'flex-start',
          fontSize: 13,
          letterSpacing: 3,
          color: hudColors.gold,
          fontWeight: 700,
          textTransform: 'uppercase',
        }}
      >
        {tr('hud.attributes')}
      </div>
      <div style={{ marginTop: 12 }}>
        <StatHexagon
          scores={scores}
          size={200}
          rings={3}
          showDots={false}
          gridStroke={gridStroke}
          fill="rgba(244,197,66,0.45)"
          fillOpacity={1}
          stroke={hudColors.gold}
          strokeWidth={2}
          labelColor="#ffffff"
          labelFont="'Pixelify Sans', sans-serif"
          labelSize={11}
        />
      </div>
    </div>
  )
}
