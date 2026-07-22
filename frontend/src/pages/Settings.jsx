import { useLanguage, LANGUAGES } from '../i18n'
import { useMusic } from '../music'
import { hudColors, roundCorners, liquidGlass, glassTextShadow } from '../hud'

// A segmented control in the HUD nav style: the selected option gets the gold
// tint + gold border used by the active nav item (see HudBar), the rest are
// translucent white. Sharp corners to keep the pixel-art feel.
function Segmented({ options, value, onChange }) {
  return (
    <div
      style={{
        display: 'inline-flex',
        flexWrap: 'wrap',
        gap: 6,
        padding: 5,
        background: 'rgba(0,0,0,0.28)',
        border: '1px solid rgba(255,255,255,0.2)',
      }}
    >
      {options.map(({ key, label }) => {
        const active = key === value
        return (
          <button
            key={key}
            type="button"
            onClick={() => onChange(key)}
            style={{
              fontFamily: 'inherit',
              fontSize: 14,
              fontWeight: active ? 700 : 400,
              padding: '8px 18px',
              cursor: 'pointer',
              color: active ? '#ffffff' : 'rgba(255,255,255,0.85)',
              background: active ? 'rgba(244,197,66,0.3)' : 'transparent',
              border: active ? `1px solid ${hudColors.gold}` : '1px solid transparent',
              textShadow: glassTextShadow,
              transition: 'background .08s',
            }}
          >
            {label}
          </button>
        )
      })}
    </div>
  )
}

// A titled setting panel: a frosty liquid-glass surface with a gold uppercase
// header (matching the ATTRIBUTES panel / START QUEST label) over white body.
function Card({ title, hint, children }) {
  return (
    <section
      style={{
        ...liquidGlass,
        ...roundCorners,
        padding: 20,
        display: 'flex',
        flexDirection: 'column',
        gap: 14,
      }}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 5 }}>
        <h2
          style={{
            margin: 0,
            fontSize: 15,
            fontWeight: 700,
            letterSpacing: 2,
            textTransform: 'uppercase',
            color: hudColors.gold,
            textShadow: glassTextShadow,
          }}
        >
          {title}
        </h2>
        <p style={{ margin: 0, fontSize: 13.5, color: 'rgba(255,255,255,0.85)', textShadow: glassTextShadow }}>{hint}</p>
      </div>
      {children}
    </section>
  )
}

// A 0–100 volume slider in the HUD palette. `value` is a 0.0–1.0 gain; onChange
// emits the same. accent-color paints the native track/thumb gold.
function VolumeSlider({ value, onChange }) {
  const pct = Math.round(value * 100)
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
      <input
        type="range"
        min={0}
        max={100}
        step={1}
        value={pct}
        onChange={(e) => onChange(Number(e.target.value) / 100)}
        aria-label="Music volume"
        style={{ flex: 1, accentColor: hudColors.gold, cursor: 'pointer' }}
      />
      <span
        style={{
          minWidth: 46,
          textAlign: 'right',
          fontFamily: "'Lowres Pixel', sans-serif",
          fontVariantNumeric: 'tabular-nums',
          fontSize: 15,
          fontWeight: 700,
          color: hudColors.gold,
          textShadow: glassTextShadow,
        }}
      >
        {pct}%
      </span>
    </div>
  )
}

export default function Settings() {
  const { lang, setLang, t: tr } = useLanguage()
  const { volume, setVolume } = useMusic()

  const langOptions = LANGUAGES.map((l) => ({ key: l.code, label: l.label }))

  return (
    <div style={{ width: '100%', maxWidth: 560, display: 'flex', flexDirection: 'column', gap: 22 }}>
      <Card title={tr('settings.language')} hint={tr('settings.language.hint')}>
        <Segmented options={langOptions} value={lang} onChange={setLang} />
      </Card>

      <Card title={tr('settings.music')} hint={tr('settings.music.hint')}>
        <VolumeSlider value={volume} onChange={setVolume} />
      </Card>
    </div>
  )
}
