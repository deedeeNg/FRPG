import { pixelated } from '../hud'
import bgDay from '../assets/hud/bg-day.png'
import bgDusk from '../assets/hud/bg-dusk.png'
import bgNight from '../assets/hud/bg-night.png'
import bgDawn from '../assets/hud/bg-dawn.png'

// Animated pixel-art landscape behind the HUD. Four full-bleed layers (day →
// dusk → night → dawn) share one silhouette, so cross-fading them on a 30s loop
// reads as time passing, not a scene change. Every layer also scrolls left by
// exactly one tile per 45s loop, so the world ambles past the character.
// See background_package/README.md for the animation spec.

// Tile width = 240/135 × layer height. The container is fixed to the viewport,
// so its height is 100vh and the fraction must be kept exact — rounding causes a
// visible seam jump when the scroll loop restarts.
const TILE_W = 'calc(100vh * 240 / 135)'

// bottom → top; each layer fades on its own phase of the shared 30s cycle.
const LAYERS = [
  { img: bgDay, fade: 'frpgFadeDay' },
  { img: bgDusk, fade: 'frpgFadeDusk' },
  { img: bgNight, fade: 'frpgFadeNight' },
  { img: bgDawn, fade: 'frpgFadeDawn' },
]

export default function Background() {
  return (
    <div
      aria-hidden
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 0,
        overflow: 'hidden',
        pointerEvents: 'none',
      }}
    >
      <style>{`
        /* day is the always-opaque base so a transition never reveals the dark
           page beneath. Each later phase fades IN on top of a still-opaque lower
           layer, then lower layers drop back to 0 only while fully hidden under
           the covering layer — so no two layers are ever both semi-transparent
           at once (which caused the mid-transition flash). */
        @keyframes frpgFadeDay   { 0%{opacity:1} 100%{opacity:1} }
        @keyframes frpgFadeDusk  { 0%{opacity:0} 20%{opacity:0} 25%{opacity:1} 90%{opacity:1} 95%{opacity:0} 100%{opacity:0} }
        @keyframes frpgFadeNight { 0%{opacity:0} 45%{opacity:0} 50%{opacity:1} 90%{opacity:1} 95%{opacity:0} 100%{opacity:0} }
        @keyframes frpgFadeDawn  { 0%{opacity:0} 70%{opacity:0} 75%{opacity:1} 95%{opacity:1} 100%{opacity:0} }
        @keyframes frpgScroll    { from { background-position-x: 0 } to { background-position-x: calc(-1 * ${TILE_W}) } }
        @media (prefers-reduced-motion: reduce) {
          .frpg-bg-layer { animation: none !important; }
        }
      `}</style>
      {LAYERS.map(({ img, fade }) => (
        <div
          key={fade}
          className="frpg-bg-layer"
          style={{
            position: 'absolute',
            inset: 0,
            backgroundImage: `url(${img})`,
            backgroundSize: 'auto 100%', // tile height = viewport height
            backgroundRepeat: 'repeat-x', // seamless horizontal tiling
            animation: `${fade} 30s linear infinite, frpgScroll 45s linear infinite`,
            ...pixelated,
          }}
        />
      ))}
    </div>
  )
}
