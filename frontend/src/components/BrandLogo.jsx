import { pixelated } from '../hud'
import fleur from '../assets/hud/fleur.png'

// Keeps the white wordmark legible over the frosty, low-tint glass — same value
// the HUD bar uses.
const defaultShadow = '0 2px 0 rgba(43,36,64,0.55)'

/**
 * FRPG brand mark: the gold pixel fleur + the "FRPG" wordmark in the Jacquard 12
 * blackletter display font. This is the mark shown in the top HUD bar; extracted
 * here so the HUD and the auth screens render the exact same logo.
 *
 * Sizing/color adapt per surface (bigger and centered on the auth cards, HUD
 * defaults on the bar) while the fleur asset and font stay fixed.
 *   fleurSize   — rendered fleur px (square)
 *   fontSize    — wordmark px
 *   letterSpacing, gap, color, textShadow
 */
export default function BrandLogo({
  fleurSize = 26,
  fontSize = 32,
  letterSpacing = 4,
  gap = 10,
  color = '#ffffff',
  textShadow = defaultShadow,
}) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap }}>
      <img src={fleur} alt="" style={{ ...pixelated, width: fleurSize, height: fleurSize }} />
      <span style={{ fontFamily: "'Jacquard 12', serif", fontSize, color, letterSpacing, textShadow }}>FRPG</span>
    </div>
  )
}
