import { useTheme } from '../theme'
import { useLanguage } from '../i18n'
import { attributes, attributeScores, MAX_SCORE } from '../data/stats'

/**
 * Character attribute hexagon (radar) chart.
 *   scores: { [attrKey]: number 0..MAX_SCORE }  — defaults to the seeded values.
 * Renders the 6 fixed axes from `attributes` as a hexagon, with the current
 * scores plotted as a filled polygon inside a set of concentric grid rings.
 */
export default function StatHexagon({ scores = attributeScores, size = 300 }) {
  const { theme: t } = useTheme()
  const { t: tr } = useLanguage()

  // The hexagon is square, but side labels ("Vocabulary", "Grammar") stick out
  // horizontally, so give the viewBox extra width on the left/right.
  const padX = size * 0.24
  const width = size + padX * 2
  const cx = width / 2
  const cy = size / 2
  const radius = size * 0.34 // leave room for labels around the edge
  const rings = 4

  // Vertex angles: start at the top (−90°) and step 60° clockwise.
  const angleFor = (i) => (-90 + i * (360 / attributes.length)) * (Math.PI / 180)
  const pointAt = (i, r) => ({
    x: cx + r * Math.cos(angleFor(i)),
    y: cy + r * Math.sin(angleFor(i)),
  })

  const toPoly = (pts) => pts.map((p) => `${p.x},${p.y}`).join(' ')

  // Outer hexagon vertices and the concentric grid rings.
  const outer = attributes.map((_, i) => pointAt(i, radius))
  const gridRings = Array.from({ length: rings }, (_, r) =>
    attributes.map((_, i) => pointAt(i, (radius * (r + 1)) / rings)),
  )

  // Score polygon.
  const scorePts = attributes.map((a, i) => {
    const v = Math.max(0, Math.min(MAX_SCORE, scores[a.key] ?? 0))
    return pointAt(i, (radius * v) / MAX_SCORE)
  })

  return (
    <svg
      width={width}
      height={size}
      viewBox={`0 0 ${width} ${size}`}
      role="img"
      aria-label={tr('stats.title')}
      style={{ display: 'block', maxWidth: '100%' }}
    >
      {/* Grid rings */}
      {gridRings.map((ring, r) => (
        <polygon
          key={r}
          points={toPoly(ring)}
          fill="none"
          stroke={t.border}
          strokeWidth={1}
        />
      ))}

      {/* Spokes */}
      {outer.map((p, i) => (
        <line key={i} x1={cx} y1={cy} x2={p.x} y2={p.y} stroke={t.border} strokeWidth={1} />
      ))}

      {/* Score area */}
      <polygon
        points={toPoly(scorePts)}
        fill={t.primary}
        fillOpacity={0.22}
        stroke={t.primary}
        strokeWidth={2}
        strokeLinejoin="round"
      />
      {scorePts.map((p, i) => (
        <circle key={i} cx={p.x} cy={p.y} r={3} fill={t.primary} />
      ))}

      {/* Axis labels */}
      {attributes.map((a, i) => {
        const p = pointAt(i, radius + 18)
        const anchor = Math.abs(p.x - cx) < 1 ? 'middle' : p.x > cx ? 'start' : 'end'
        return (
          <text
            key={a.key}
            x={p.x}
            y={p.y}
            textAnchor={anchor}
            dominantBaseline="middle"
            fontSize={16}
            fontWeight={700}
            fill={t.soft}
            style={{ fontFamily: "'Public Sans', system-ui, sans-serif" }}
          >
            {tr(a.labelKey)}
          </text>
        )
      })}
    </svg>
  )
}
