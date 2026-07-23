import { useState } from 'react'
import { useTheme } from '../theme'

export default function TextField({
  label,
  type = 'text',
  name,
  placeholder = '',
  autoComplete = 'off',
  value,
  onChange,
  gap = 16,
  // Override the label styling (e.g. white text on a dark "liquid glass" card).
  labelColor,
  labelShadow,
}) {
  const { theme: t } = useTheme()
  const [focused, setFocused] = useState(false)

  const input = {
    width: '100%',
    boxSizing: 'border-box',
    fontFamily: 'inherit',
    fontSize: 15,
    color: t.ink,
    background: focused ? t.surface : t.field,
    border: `1px solid ${focused ? t.primary : t.fieldBorder}`,
    borderRadius: 12,
    padding: '13px 15px',
    outline: 'none',
    boxShadow: focused ? `0 0 0 3px ${t.ring}` : 'none',
    transition: 'background .15s, border-color .15s, box-shadow .15s',
  }

  return (
    <label style={{ display: 'flex', flexDirection: 'column', gap: 7, marginBottom: gap }}>
      <span style={{ fontSize: 12.5, fontWeight: 600, color: labelColor ?? t.labelColor, letterSpacing: '.01em', textShadow: labelShadow }}>{label}</span>
      <input
        type={type}
        name={name}
        placeholder={placeholder}
        autoComplete={autoComplete}
        value={value}
        onChange={onChange}
        onFocus={() => setFocused(true)}
        onBlur={() => setFocused(false)}
        style={input}
      />
    </label>
  )
}
