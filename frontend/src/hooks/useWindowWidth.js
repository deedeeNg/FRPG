import { useEffect, useState } from 'react'

// Current viewport width, updated on resize. Used to switch the HUD between its
// pixel-perfect absolute desktop layout and a stacked flow layout on narrow
// screens.
export function useWindowWidth() {
  const [width, setWidth] = useState(() => (typeof window === 'undefined' ? 1440 : window.innerWidth))

  useEffect(() => {
    const onResize = () => setWidth(window.innerWidth)
    window.addEventListener('resize', onResize)
    return () => window.removeEventListener('resize', onResize)
  }, [])

  return width
}

// Below this width the absolute desktop layout (heading left, radar right) would
// overlap, so the HUD reflows into a single stacked column.
export const HUD_BREAKPOINT = 1100

// Below this width the quest cards collapse to icon + title (subtitle and
// "Start Quest" hidden until the card is expanded).
export const CARD_COLLAPSE_BREAKPOINT = 640
