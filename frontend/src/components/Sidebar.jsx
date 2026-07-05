import { useState } from 'react'
import { useTheme } from '../theme'

// shadcn sidebar widths: --sidebar-width 16rem, --sidebar-width-icon ~3.5rem.
const WIDTH = 256
const WIDTH_ICON = 56

// lucide-style icons (24x24, stroke = currentColor) to match shadcn's sidebar.
const Icon = ({ children, size = 16 }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
    style={{ flex: '0 0 auto' }}
  >
    {children}
  </svg>
)

const HomeIcon = () => (
  <Icon>
    <path d="M15 21v-6a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1v6" />
    <path d="M3 10a2 2 0 0 1 .709-1.528l7-5.999a2 2 0 0 1 2.582 0l7 5.999A2 2 0 0 1 21 10v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
  </Icon>
)

const MapIcon = () => (
  <Icon>
    <path d="M14.106 5.553a2 2 0 0 0 1.788 0l3.659-1.83A1 1 0 0 1 21 4.619v12.764a1 1 0 0 1-.553.894l-4.553 2.277a2 2 0 0 1-1.788 0l-4.212-2.106a2 2 0 0 0-1.788 0l-3.659 1.83A1 1 0 0 1 3 19.381V6.618a1 1 0 0 1 .553-.894l4.553-2.277a2 2 0 0 1 1.788 0z" />
    <path d="M15 5.764v15" />
    <path d="M9 3.236v15" />
  </Icon>
)

const LearningIcon = () => (
  <Icon>
    <path d="M12 7v14" />
    <path d="M3 18a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1h5a4 4 0 0 1 4 4 4 4 0 0 1 4-4h5a1 1 0 0 1 1 1v13a1 1 0 0 1-1 1h-6a3 3 0 0 0-3 3 3 3 0 0 0-3-3z" />
  </Icon>
)

const SettingsIcon = () => (
  <Icon>
    <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" />
    <circle cx="12" cy="12" r="3" />
  </Icon>
)

const LogOutIcon = () => (
  <Icon>
    <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
    <polyline points="16 17 21 12 16 7" />
    <line x1="21" x2="9" y1="12" y2="12" />
  </Icon>
)

const TABS = [
  { key: 'home', label: 'Home', Glyph: HomeIcon },
  { key: 'map', label: 'Map', Glyph: MapIcon },
  { key: 'learning', label: 'Learning', Glyph: LearningIcon },
  { key: 'settings', label: 'Settings', Glyph: SettingsIcon },
]

// shadcn-style side navbar. Collapses between full width and an icon-only rail;
// `open` / `onToggle` are owned by the app shell (App.jsx) alongside the trigger,
// keyboard shortcut (cmd/ctrl+b) and cookie persistence.
export default function Sidebar({ open, onToggle, active, onSelect, onLogout }) {
  const { theme: t } = useTheme()
  const [hover, setHover] = useState(null)

  // shadcn SidebarMenuButton: rounded row, icon + label, hover/active = accent bg.
  // Collapsed (icon) mode centers the icon and drops the label.
  const item = (isActive, isHover) => ({
    display: 'flex',
    alignItems: 'center',
    justifyContent: open ? 'flex-start' : 'center',
    gap: 10,
    width: '100%',
    height: 34,
    textAlign: 'left',
    fontFamily: 'inherit',
    fontSize: 14,
    fontWeight: isActive ? 600 : 500,
    border: 'none',
    borderRadius: 8,
    padding: open ? '0 10px' : 0,
    cursor: 'pointer',
    whiteSpace: 'nowrap',
    overflow: 'hidden',
    transition: 'background .15s, color .15s',
    background: isActive || isHover ? t.field : 'transparent',
    color: isActive ? t.ink : t.soft,
  })

  return (
    <div
      style={{
        position: 'relative',
        flex: `0 0 ${open ? WIDTH : WIDTH_ICON}px`,
        width: open ? WIDTH : WIDTH_ICON,
        height: '100%',
        overflowY: 'auto',
        overflowX: 'hidden',
        background: t.surface,
        borderRight: `1px solid ${t.border}`,
        boxSizing: 'border-box',
        fontFamily: "'Bricolage Grotesque', sans-serif",
        display: 'flex',
        flexDirection: 'column',
        transition: 'flex-basis .2s ease, width .2s ease',
      }}
    >
      {/* Header: Gmail-style hamburger toggle + brand. Collapsed shows only the
          hamburger, centered. */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: open ? 'flex-start' : 'center',
          gap: 8,
          padding: open ? '12px 10px' : '12px 0',
          borderBottom: `1px solid ${t.border}`,
        }}
      >
        <button
          type="button"
          aria-label="Main menu"
          title="Main menu"
          onClick={onToggle}
          onMouseEnter={() => setHover('menu')}
          onMouseLeave={() => setHover(null)}
          style={{
            width: 40,
            height: 40,
            flex: '0 0 auto',
            borderRadius: '50%',
            border: 'none',
            background: hover === 'menu' ? t.field : 'transparent',
            color: t.soft,
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            padding: 0,
            transition: 'background .15s',
          }}
        >
          <Icon size={20}>
            <line x1="4" x2="20" y1="6" y2="6" />
            <line x1="4" x2="20" y1="12" y2="12" />
            <line x1="4" x2="20" y1="18" y2="18" />
          </Icon>
        </button>
        {open && (
          <>
            <div
              style={{
                width: 30,
                height: 30,
                flex: '0 0 auto',
                borderRadius: 8,
                background: t.primary,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: '#fff',
                fontWeight: 800,
                fontSize: 13,
                letterSpacing: '.02em',
              }}
            >
              F
            </div>
            <span style={{ fontWeight: 700, fontSize: 14, color: t.ink, whiteSpace: 'nowrap' }}>FRPG</span>
          </>
        )}
      </div>

      {/* Nav group */}
      <div style={{ padding: open ? '12px 8px' : '12px 8px', display: 'flex', flexDirection: 'column', gap: 2 }}>
        {open && (
          <div
            style={{
              fontSize: 11,
              fontWeight: 600,
              textTransform: 'uppercase',
              letterSpacing: '.06em',
              color: t.faint,
              padding: '4px 10px 6px',
            }}
          >
            Platform
          </div>
        )}
        {TABS.map(({ key, label, Glyph }) => (
          <button
            key={key}
            type="button"
            title={open ? undefined : label}
            onClick={() => onSelect(key)}
            onMouseEnter={() => setHover(key)}
            onMouseLeave={() => setHover(null)}
            style={item(active === key, hover === key)}
          >
            <Glyph />
            {open && label}
          </button>
        ))}
      </div>

      {/* Footer */}
      <div style={{ marginTop: 'auto', padding: 8, borderTop: `1px solid ${t.border}` }}>
        <button
          type="button"
          title={open ? undefined : 'Log Out'}
          onClick={onLogout}
          onMouseEnter={() => setHover('logout')}
          onMouseLeave={() => setHover(null)}
          style={item(false, hover === 'logout')}
        >
          <LogOutIcon />
          {open && 'Log Out'}
        </button>
      </div>

    </div>
  )
}
