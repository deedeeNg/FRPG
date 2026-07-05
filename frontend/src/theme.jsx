import { createContext, useContext, useEffect, useState } from 'react'

// ---------------------------------------------------------------------------
// Single source of truth for all colors. Reskin the whole app by editing here.
// ---------------------------------------------------------------------------
export const lightTheme = {
  page: '#F5F0E7', surface: '#FFFFFF', border: '#E7E0D5',
  ink: '#221E2E', soft: '#6B6478', mute: '#8A8395', faint: '#9A93A2',
  field: '#F5F1EA', fieldBorder: '#E4DDD0', labelColor: '#4A4553',
  primary: '#4A3FB5', primaryHover: '#3E349A', ring: 'rgba(74,63,181,.14)', gemGlow: 'rgba(74,63,181,.7)',
  divider: '#E7E0D5',
  socialBg: '#FFFFFF', socialBorder: '#E0D8CB', socialHoverBg: '#FAF7F1', socialHoverBorder: '#CFC6B7',
  mono: '#F1EEE8', monoText: '#4A4553',
  tabWrap: '#EBE4D8', tabWrapBorder: '#E1D9CB', tabActiveBg: '#FFFFFF', tabActiveText: '#4A3FB5', tabShadow: 'rgba(34,30,46,.14)', tabIdle: '#8A8395',
  heroA: '#EDE7DC', heroB: '#F4EFE6', heroLabelBg: '#F5F0E7', heroLabelBorder: '#CFC6B7', heroLabelText: '#9A9083',
  skillHoverBorder: '#CBBFF0', skillShadow: 'rgba(74,63,181,.5)', tokenWrap: '#F3F1FB',
  gem: '#E7C24B', icon: '#4A4553',
  cardShadow: '0 1px 2px rgba(34,30,46,.04),0 18px 40px -24px rgba(34,30,46,.22)',
}

export const darkTheme = {
  page: '#17151F', surface: '#221F2C', border: '#322E3D',
  ink: '#F1EEF6', soft: '#A69FB4', mute: '#8F88A0', faint: '#7C7690',
  field: '#1B1925', fieldBorder: '#37333F', labelColor: '#C6BFD3',
  primary: '#6F62D9', primaryHover: '#7E72E4', ring: 'rgba(111,98,217,.30)', gemGlow: 'rgba(111,98,217,.55)',
  divider: '#322E3D',
  socialBg: '#221F2C', socialBorder: '#3A3547', socialHoverBg: '#2A2636', socialHoverBorder: '#4A4459',
  mono: '#2C2838', monoText: '#C9C2D6',
  tabWrap: '#221F2C', tabWrapBorder: '#322E3D', tabActiveBg: '#3A3550', tabActiveText: '#CBC3F5', tabShadow: 'rgba(0,0,0,.4)', tabIdle: '#8F88A0',
  heroA: '#26232F', heroB: '#2D2A38', heroLabelBg: '#17151F', heroLabelBorder: '#444050', heroLabelText: '#8F88A0',
  skillHoverBorder: '#5B4FC4', skillShadow: 'rgba(0,0,0,.55)', tokenWrap: '#2A2636',
  gem: '#E7C24B', icon: '#E7C24B',
  cardShadow: '0 1px 2px rgba(0,0,0,.3),0 20px 44px -24px rgba(0,0,0,.7)',
}

const STORAGE_KEY = 'frpg-theme'

// Theme mode: an explicit 'light'/'dark' choice, or 'system' to follow the OS
// via prefers-color-scheme. `dark` below is the resolved boolean.
export const THEME_MODES = ['light', 'dark', 'system']

function systemPrefersDark() {
  return typeof window !== 'undefined' && !!window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
}

const ThemeContext = createContext({ theme: lightTheme, dark: false, mode: 'system', setMode: () => {}, toggle: () => {} })

export function ThemeProvider({ children, defaultMode = 'system' }) {
  const [mode, setMode] = useState(() => {
    if (typeof window === 'undefined') return defaultMode
    const saved = window.localStorage.getItem(STORAGE_KEY)
    return THEME_MODES.includes(saved) ? saved : defaultMode
  })

  // Track the OS preference so 'system' mode reacts live.
  const [systemDark, setSystemDark] = useState(systemPrefersDark)
  useEffect(() => {
    if (typeof window === 'undefined' || !window.matchMedia) return
    const mq = window.matchMedia('(prefers-color-scheme: dark)')
    const onChange = (e) => setSystemDark(e.matches)
    mq.addEventListener('change', onChange)
    return () => mq.removeEventListener('change', onChange)
  }, [])

  useEffect(() => {
    window.localStorage.setItem(STORAGE_KEY, mode)
  }, [mode])

  const dark = mode === 'system' ? systemDark : mode === 'dark'

  const value = {
    theme: dark ? darkTheme : lightTheme,
    dark,
    mode,
    setMode,
    // Quick top-right switch: pin to the opposite of what's showing now.
    toggle: () => setMode(dark ? 'light' : 'dark'),
  }

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}

export function useTheme() {
  return useContext(ThemeContext)
}
