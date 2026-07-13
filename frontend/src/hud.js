// Shared style tokens for the pixel-art RPG HUD home screen (design "2a — HUD
// Aventure"). Colors, the pixel-cut corner clip-path, and the dark "liquid
// glass" surface are used across the HUD bar, quest cards, and attributes panel.
// See design_handoff_frpg_pixel_rpg_home/README.md for the full spec.

// Color tokens (see README "Design Tokens").
export const hudColors = {
  ink: '#2b2440', // navy shadow base; shadows use rgba(43,36,64,…)
  gold: '#f4c542', // accent / XP / active nav
  goldLight: '#ffdf6b', // XP stripe highlight
  blue: '#3a6ec8', // Speaking
  green: '#2f9e5b', // Reading
  rose: '#e85d75', // Writing
}

// Rounded card corners.
export const roundCorners = {
  borderRadius: 16,
}

// Dark translucent panel with backdrop blur, hairline border, inset highlight.
export const glassDark = {
  background: 'rgba(28,22,52,0.55)',
  backdropFilter: 'blur(14px) saturate(1.4)',
  WebkitBackdropFilter: 'blur(14px) saturate(1.4)',
  border: '1px solid rgba(255,255,255,0.25)',
  boxShadow: 'inset 0 1px 0 rgba(255,255,255,0.2)',
}

// Lighter, frostier "liquid glass": much more translucent than glassDark with a
// heavier blur, so the landscape reads clearly through it. Used by quest cards.
export const liquidGlass = {
  background: 'rgba(28,22,52,0.28)',
  backdropFilter: 'blur(22px) saturate(1.8)',
  WebkitBackdropFilter: 'blur(22px) saturate(1.8)',
  border: '1px solid rgba(255,255,255,0.35)',
  boxShadow: 'inset 0 1px 0 rgba(255,255,255,0.35)',
}
// Hover fill for liquid-glass surfaces (kept translucent, just a touch denser).
export const liquidGlassHover = 'rgba(28,22,52,0.42)'

// Pixel art must never smooth-scale.
export const pixelated = { imageRendering: 'pixelated' }

// Keeps white text legible over the frosty, low-tint liquid glass.
export const glassTextShadow = '0 2px 0 rgba(43,36,64,0.55)'
