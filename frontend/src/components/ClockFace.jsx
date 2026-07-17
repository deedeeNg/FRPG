import { useId } from 'react'
import { hudColors } from '../hud'

// Renders a clock face for a { hour, minute } exercise stimulus (24h input; the
// dial itself is always a plain 12h face — hour%12 drives the hour-hand angle
// regardless of which French register the question is testing).
//
// A 12h dial alone can't distinguish e.g. 11:30 from 23:30 — both render
// identical hand positions, yet the formal register's correct answer differs
// ("onze heures trente" vs "vingt-trois heures trente"). A sun/moon badge
// (hour 0-11 = AM/sun, hour 12-23 = PM/moon — the standard AM/PM split, not a
// literal daylight guess) resolves that.
//
// Layout: the badge sits ABOVE the dial, both on the vertical centerline, and the
// [badge | gap | dial] stack is vertically centered in the square canvas — so the
// two elements read as one centered composition with clear separation, rather than
// a centered dial with a badge stuck to a corner.
export default function ClockFace({ hour, minute, size = 150 }) {
  const cx = size / 2 // both badge and dial share this vertical centerline
  const R = size * 0.32 // dial radius
  const badgeR = size * 0.08
  const badgeOuter = badgeR * 1.5 // generous bound (sun rays + stroke) for layout
  const gap = size * 0.09 // clear space between the badge and the dial ring

  // Vertically center the [badge | gap | dial] stack within the square, then nudge
  // the badge right (~3rem base + 10% of size) and down (5% of size) off that
  // centered position. Both nudges are clamped so the badge's outer edge never
  // clips the canvas.
  const stackH = 2 * badgeOuter + gap + 2 * R
  const top = (size - stackH) / 2
  const dialCy = top + 2 * badgeOuter + gap + R
  const badgeCx = Math.min(cx + 48 + size * 0.1, size - badgeOuter)
  const badgeCy = Math.min(top + badgeOuter + size * 0.05, dialCy - R - badgeOuter - 1.5)

  const hourAngle = ((hour % 12) + minute / 60) * 30
  const minuteAngle = minute * 6
  const isPM = hour >= 12
  const maskId = useId()

  // Points on the dial, measured from its center (cx, dialCy).
  const point = (angleDeg, radius) => {
    const rad = ((angleDeg - 90) * Math.PI) / 180
    return [cx + radius * Math.cos(rad), dialCy + radius * Math.sin(rad)]
  }

  const hand = (angleDeg, length, width, color) => {
    const [x2, y2] = point(angleDeg, length)
    return <line x1={cx} y1={dialCy} x2={x2} y2={y2} stroke={color} strokeWidth={width} strokeLinecap="round" />
  }

  const ticks = Array.from({ length: 12 }, (_, i) => {
    const [x1, y1] = point(i * 30, i % 3 === 0 ? R - 11 : R - 8)
    const [x2, y2] = point(i * 30, R - 4)
    return <line key={i} x1={x1} y1={y1} x2={x2} y2={y2} stroke="rgba(255,255,255,0.6)" strokeWidth={i % 3 === 0 ? 2.5 : 1.5} />
  })

  // Sun: filled circle + 8 rays. Moon: crescent, cut via a mask (one circle minus
  // an offset copy). Both drawn at the origin, then translated to the badge center.
  const sun = (
    <g>
      <circle r={badgeR * 0.7} fill={hudColors.gold} />
      {Array.from({ length: 8 }, (_, i) => {
        const rad = (i * 45 * Math.PI) / 180
        const r1 = badgeR * 0.95
        const r2 = badgeR * 1.35
        return (
          <line
            key={i}
            x1={r1 * Math.cos(rad)}
            y1={r1 * Math.sin(rad)}
            x2={r2 * Math.cos(rad)}
            y2={r2 * Math.sin(rad)}
            stroke={hudColors.gold}
            strokeWidth={badgeR * 0.22}
            strokeLinecap="round"
          />
        )
      })}
    </g>
  )
  const moon = (
    <g>
      <mask id={maskId}>
        <rect x={-badgeR * 1.5} y={-badgeR * 1.5} width={badgeR * 3} height={badgeR * 3} fill="white" />
        <circle cx={badgeR * 0.45} cy={-badgeR * 0.15} r={badgeR * 0.85} fill="black" />
      </mask>
      <circle r={badgeR} fill="#ffffff" mask={`url(#${maskId})`} />
    </g>
  )

  return (
    <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`} style={{ flex: 'none' }}>
      <g transform={`translate(${badgeCx}, ${badgeCy})`}>{isPM ? moon : sun}</g>
      <circle cx={cx} cy={dialCy} r={R} fill="rgba(28,22,52,0.4)" stroke={hudColors.gold} strokeWidth={3} />
      {ticks}
      {hand(hourAngle, R * 0.5, 4, '#ffffff')}
      {hand(minuteAngle, R * 0.75, 2.5, hudColors.gold)}
      <circle cx={cx} cy={dialCy} r={4} fill={hudColors.gold} />
    </svg>
  )
}
