import HudBar from './HudBar'
import Background from './Background'
import { pixelated } from '../hud'
import { userProfile } from '../data/user'
import { useWindowWidth, HUD_BREAKPOINT } from '../hooks/useWindowWidth'

/**
 * Full-bleed HUD shell: the pixel-art landscape background with the top HUD bar.
 * On desktop the bar is anchored 20px from the top/left/right and screens lay
 * their content out absolutely over the landscape. Below {@link HUD_BREAKPOINT}
 * the layout reflows: the bar sits in normal flow at the top and content stacks
 * beneath it in a scrollable column.
 *   activeRoute, onNavigate(route), onLogout, showXpBar
 */
export default function HudLayout({ activeRoute, onNavigate, onLogout, showXpBar = true, children }) {
  const width = useWindowWidth()
  const compact = width < HUD_BREAKPOINT

  return (
    <div
      style={{
        position: 'relative',
        width: '100%',
        minHeight: '100vh',
        overflowX: 'hidden',
        overflowY: 'auto',
        fontFamily: "'Pixelify Sans', sans-serif",
        ...pixelated,
      }}
    >
      <Background />
      <div
        style={
          compact
            ? { position: 'relative', margin: 20 }
            : { position: 'absolute', left: 20, right: 20, top: 20 }
        }
      >
        <HudBar
          user={userProfile}
          activeRoute={activeRoute}
          onNavigate={onNavigate}
          onLogout={onLogout}
          showXpBar={showXpBar}
          compact={compact}
        />
      </div>
      {children}
    </div>
  )
}
