import { createContext, useContext, useEffect, useRef, useState } from 'react'
import track from './assets/audio/la-vie-en-rose.mp3'

// ---------------------------------------------------------------------------
// Background music: a single, app-wide looping <audio> instance plus its volume
// control. Lives above the screen switcher (see App) so play state and volume
// survive route changes — there is exactly ONE audio element for the whole app,
// never one per screen.
//
// Autoplay policy: browsers block audio until the user interacts with the page.
// We don't fight it — we arm one-time pointer/key listeners and start playback
// on the first user gesture, then drop the listeners.
//
// Volume is persisted to localStorage ('frpg-music-volume', a 0.0–1.0 gain),
// consistent with how map progress/theme/language persist. 0 = muted.
// ---------------------------------------------------------------------------

const STORAGE_KEY = 'frpg-music-volume'
const DEFAULT_VOLUME = 0.5

function readVolume() {
  if (typeof window === 'undefined') return DEFAULT_VOLUME
  const raw = window.localStorage.getItem(STORAGE_KEY)
  const n = raw == null ? NaN : Number(raw)
  return Number.isFinite(n) ? Math.min(1, Math.max(0, n)) : DEFAULT_VOLUME
}

const MusicContext = createContext({ volume: DEFAULT_VOLUME, setVolume: () => {} })

export function MusicProvider({ children }) {
  const [volume, setVolume] = useState(readVolume)
  const audioRef = useRef(null)

  // Create the single audio element once.
  if (audioRef.current === null && typeof Audio !== 'undefined') {
    const el = new Audio(track)
    el.loop = true
    el.preload = 'auto'
    el.volume = readVolume()
    audioRef.current = el
  }

  // Arm autoplay on the first user interaction, then disarm.
  useEffect(() => {
    const el = audioRef.current
    if (!el) return
    const start = () => {
      el.play().catch(() => {}) // still blocked (e.g. no gesture yet) — ignore.
      remove()
    }
    const remove = () => {
      window.removeEventListener('pointerdown', start)
      window.removeEventListener('keydown', start)
    }
    window.addEventListener('pointerdown', start)
    window.addEventListener('keydown', start)
    return remove
  }, [])

  // Live-apply volume and persist it.
  useEffect(() => {
    if (audioRef.current) audioRef.current.volume = volume
    try {
      window.localStorage.setItem(STORAGE_KEY, String(volume))
    } catch {
      // Storage unavailable (private mode): volume just won't persist.
    }
  }, [volume])

  return <MusicContext.Provider value={{ volume, setVolume }}>{children}</MusicContext.Provider>
}

export function useMusic() {
  return useContext(MusicContext)
}
