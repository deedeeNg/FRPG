import { useState } from 'react'
import { useLanguage } from '../i18n'
import { hudColors, pxCorners, liquidGlass, pixelated } from '../hud'

// Keep white HUD text legible over the frosty, low-tint glass.
const textShadow = '0 2px 0 rgba(43,36,64,0.55)'
import knight from '../assets/hud/knight.png'
import fleur from '../assets/hud/fleur.png'
import iconHome from '../assets/hud/icon-home.png'
import iconMap from '../assets/hud/icon-map.png'
import iconBook from '../assets/hud/icon-book.png'
import iconGear from '../assets/hud/icon-gear.png'
import iconDoor from '../assets/hud/icon-door.png'

// Nav items, left to right. `route` drives the active state and onNavigate;
// Log Out is special-cased (ends the session instead of routing).
const NAV = [
  { route: 'home', icon: iconHome, labelKey: 'nav.home' },
  { route: 'map', icon: iconMap, labelKey: 'nav.map' },
  { route: 'learning', icon: iconBook, labelKey: 'nav.learning' },
  { route: 'settings', icon: iconGear, labelKey: 'nav.settings' },
  { route: 'logout', icon: iconDoor, labelKey: 'nav.logout' },
]

/**
 * Top HUD bar: knight avatar chip, XP block, centered fleur + FRPG logo, and
 * the nav. Fluid width; anchored 20px from the top/left/right by the caller.
 *   user: { name, level, currentXp, nextLevelXp }
 *   activeRoute, onNavigate(route), onLogout, showXpBar
 */
export default function HudBar({ user, activeRoute, onNavigate, onLogout, showXpBar = true, compact = false }) {
  const { t: tr } = useLanguage()
  const [hover, setHover] = useState(null)

  const pct = user.nextLevelXp > 0 ? Math.max(0, Math.min(1, user.currentXp / user.nextLevelXp)) : 0

  const navItem = (route, active, hovered) => ({
    display: 'flex',
    alignItems: 'center',
    gap: 8,
    padding: '9px 14px',
    fontSize: 15,
    cursor: 'pointer',
    color: active ? '#ffffff' : 'rgba(255,255,255,0.85)',
    fontWeight: active ? 600 : 400,
    background: active ? 'rgba(244,197,66,0.3)' : hovered ? 'rgba(255,255,255,0.12)' : 'transparent',
    border: active ? `1px solid ${hudColors.gold}` : '1px solid transparent',
    textShadow,
    transition: 'background .08s',
  })

  return (
    <div
      style={{
        ...liquidGlass,
        ...pxCorners,
        minHeight: 76,
        display: 'flex',
        alignItems: 'center',
        flexWrap: compact ? 'wrap' : 'nowrap',
        padding: compact ? '12px 18px' : '0 24px',
        gap: compact ? 12 : 20,
        boxSizing: 'border-box',
      }}
    >
      {/* Avatar chip: knight sprite, bottom-cropped inside a gold-bordered box. */}
      <div
        style={{
          width: 52,
          height: 52,
          flex: 'none',
          background: 'rgba(244,197,66,0.25)',
          border: `2px solid ${hudColors.gold}`,
          display: 'flex',
          alignItems: 'flex-end',
          justifyContent: 'center',
          overflow: 'hidden',
        }}
      >
        <img src={knight} alt="" style={{ ...pixelated, width: 40, marginBottom: -18 }} />
      </div>

      {/* XP block */}
      {showXpBar && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 5, flex: compact ? '1 1 200px' : '0 0 300px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 14, color: '#ffffff', fontWeight: 600, textShadow }}>
            <span>
              {user.name} · {tr('hud.level')} {user.level}
            </span>
            <span style={{ color: hudColors.gold }}>
              {user.currentXp.toLocaleString()} / {user.nextLevelXp.toLocaleString()} {tr('hud.xpSuffix')}
            </span>
          </div>
          <div style={{ height: 14, background: 'rgba(0,0,0,0.45)', border: '2px solid rgba(255,255,255,0.6)', padding: 2, boxSizing: 'border-box' }}>
            <div
              style={{
                height: '100%',
                width: `${pct * 100}%`,
                background: `repeating-linear-gradient(90deg, ${hudColors.gold} 0 8px, ${hudColors.goldLight} 8px 16px)`,
              }}
            />
          </div>
        </div>
      )}

      {/* Logo — centered on desktop; flows inline when the bar wraps. */}
      <div
        style={{
          marginLeft: compact ? 0 : 'auto',
          marginRight: compact ? 0 : 'auto',
          display: 'flex',
          alignItems: 'center',
          gap: 10,
        }}
      >
        <img src={fleur} alt="" style={{ ...pixelated, width: 26, height: 26 }} />
        <span style={{ fontFamily: "'Jacquard 12', serif", fontSize: 32, color: '#ffffff', letterSpacing: 4, textShadow }}>FRPG</span>
      </div>

      {/* Nav — wraps when compact; labels collapse to icon-only to save room. */}
      <nav style={{ display: 'flex', alignItems: 'center', flexWrap: 'wrap', gap: 8, marginLeft: compact ? 'auto' : 0 }}>
        {NAV.map(({ route, icon, labelKey }) => {
          const active = route === activeRoute
          const label = tr(labelKey)
          return (
            <div
              key={route}
              role="button"
              tabIndex={0}
              title={compact ? label : undefined}
              onClick={() => (route === 'logout' ? onLogout() : onNavigate(route))}
              onKeyDown={(e) =>
                (e.key === 'Enter' || e.key === ' ') && (route === 'logout' ? onLogout() : onNavigate(route))
              }
              onMouseEnter={() => setHover(route)}
              onMouseLeave={() => setHover(null)}
              style={navItem(route, active, hover === route)}
            >
              <img src={icon} alt="" style={{ ...pixelated, width: 16, height: 16, opacity: active ? 1 : 0.8 }} />
              {!compact && <span>{label}</span>}
            </div>
          )
        })}
      </nav>
    </div>
  )
}
