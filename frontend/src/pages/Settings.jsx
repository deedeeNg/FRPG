import { useTheme } from '../theme'
import { useLanguage, LANGUAGES } from '../i18n'

// A pill segmented control matching the app's tab styling (see App.jsx `tab`).
function Segmented({ options, value, onChange, theme: t }) {
  return (
    <div
      style={{
        display: 'inline-flex',
        flexWrap: 'wrap',
        gap: 4,
        padding: 4,
        background: t.tabWrap,
        border: `1px solid ${t.tabWrapBorder}`,
        borderRadius: 999,
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
              fontSize: 13.5,
              fontWeight: 700,
              border: 'none',
              borderRadius: 999,
              padding: '8px 18px',
              cursor: 'pointer',
              letterSpacing: '.01em',
              transition: 'background .15s, color .15s',
              background: active ? t.tabActiveBg : 'transparent',
              color: active ? t.tabActiveText : t.tabIdle,
              boxShadow: active ? `0 1px 3px ${t.tabShadow}` : 'none',
            }}
          >
            {label}
          </button>
        )
      })}
    </div>
  )
}

// A titled card wrapping one setting control.
function Card({ title, hint, children, theme: t }) {
  return (
    <section
      style={{
        background: t.surface,
        border: `1px solid ${t.border}`,
        borderRadius: 16,
        padding: 20,
        boxShadow: t.cardShadow,
        display: 'flex',
        flexDirection: 'column',
        gap: 14,
      }}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
        <h2 style={{ margin: 0, fontSize: 16, fontWeight: 700, color: t.ink }}>{title}</h2>
        <p style={{ margin: 0, fontSize: 13.5, color: t.soft }}>{hint}</p>
      </div>
      {children}
    </section>
  )
}

export default function Settings() {
  const { theme: t, mode, setMode } = useTheme()
  const { lang, setLang, t: tr } = useLanguage()

  const themeOptions = [
    { key: 'light', label: tr('settings.theme.light') },
    { key: 'dark', label: tr('settings.theme.dark') },
    { key: 'system', label: tr('settings.theme.system') },
  ]

  const langOptions = LANGUAGES.map((l) => ({ key: l.code, label: l.label }))

  return (
    <div style={{ width: '100%', maxWidth: 560, display: 'flex', flexDirection: 'column', gap: 24 }}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
        <h1 style={{ margin: 0, fontSize: 26, fontWeight: 800, color: t.ink, letterSpacing: '-.01em' }}>
          {tr('settings.title')}
        </h1>
        <p style={{ margin: 0, fontSize: 14.5, color: t.soft }}>{tr('settings.subtitle')}</p>
      </div>

      <Card title={tr('settings.appearance')} hint={tr('settings.appearance.hint')} theme={t}>
        <Segmented options={themeOptions} value={mode} onChange={setMode} theme={t} />
      </Card>

      <Card title={tr('settings.language')} hint={tr('settings.language.hint')} theme={t}>
        <Segmented options={langOptions} value={lang} onChange={setLang} theme={t} />
      </Card>
    </div>
  )
}
