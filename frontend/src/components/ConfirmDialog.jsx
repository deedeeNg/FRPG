import { useEffect, useRef } from 'react'
import { hudColors, roundCorners, glassDark, glassTextShadow } from '../hud'

// Confirmation modal in the pixel-art RPG HUD style (matches QuestionModal and
// the Legend card): dark liquid-glass panel over a dimmed backdrop (no
// outside-click dismiss), gold title, and a right-aligned footer with a ghost
// Cancel and a gold Action (rose when destructive). Escape triggers Cancel.
export default function ConfirmDialog({
  open,
  title,
  description,
  cancelLabel = 'Cancel',
  actionLabel = 'Continue',
  destructive = false,
  onCancel,
  onConfirm,
}) {
  const actionRef = useRef(null)

  // Escape cancels; move focus to the action button on open.
  useEffect(() => {
    if (!open) return
    const onKey = (e) => {
      if (e.key === 'Escape') {
        e.preventDefault()
        onCancel()
      }
    }
    window.addEventListener('keydown', onKey)
    actionRef.current?.focus()
    return () => window.removeEventListener('keydown', onKey)
  }, [open, onCancel])

  if (!open) return null

  const actionBg = destructive ? hudColors.rose : hudColors.gold
  const actionInk = destructive ? '#ffffff' : hudColors.ink

  return (
    <div
      role="alertdialog"
      aria-modal="true"
      aria-labelledby="confirm-title"
      aria-describedby="confirm-desc"
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 50,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 20,
        fontFamily: "'Pixelify Sans', sans-serif",
        background: 'rgba(10,8,20,0.55)',
        backdropFilter: 'blur(2px)',
        WebkitBackdropFilter: 'blur(2px)',
        animation: 'frpgDialogFade .15s ease',
      }}
    >
      <style>{`
        @keyframes frpgDialogFade { from { opacity: 0 } to { opacity: 1 } }
        @keyframes frpgDialogZoom { from { opacity: 0; transform: scale(.96) } to { opacity: 1; transform: scale(1) } }
      `}</style>
      <div
        style={{
          ...glassDark,
          ...roundCorners,
          width: '100%',
          maxWidth: 440,
          padding: 24,
          color: '#ffffff',
          animation: 'frpgDialogZoom .15s ease',
        }}
      >
        {/* Header */}
        <h2
          id="confirm-title"
          style={{ margin: 0, fontSize: 20, fontWeight: 700, color: hudColors.gold, textShadow: glassTextShadow }}
        >
          {title}
        </h2>
        {description && (
          <p
            id="confirm-desc"
            style={{ margin: '10px 0 0', fontSize: 15, lineHeight: 1.5, color: '#ffffff', textShadow: glassTextShadow }}
          >
            {description}
          </p>
        )}

        {/* Footer */}
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 10, marginTop: 24 }}>
          <button
            type="button"
            onClick={onCancel}
            onMouseEnter={(e) => (e.currentTarget.style.background = 'rgba(255,255,255,0.16)')}
            onMouseLeave={(e) => (e.currentTarget.style.background = 'rgba(255,255,255,0.08)')}
            style={{
              fontFamily: 'inherit',
              fontSize: 15,
              fontWeight: 700,
              padding: '11px 20px',
              borderRadius: 10,
              cursor: 'pointer',
              background: 'rgba(255,255,255,0.08)',
              border: '1.5px solid rgba(255,255,255,0.3)',
              color: '#ffffff',
              textShadow: glassTextShadow,
              transition: 'background .08s',
            }}
          >
            {cancelLabel}
          </button>
          <button
            ref={actionRef}
            type="button"
            onClick={onConfirm}
            style={{
              fontFamily: 'inherit',
              fontSize: 15,
              fontWeight: 700,
              padding: '11px 22px',
              borderRadius: 10,
              cursor: 'pointer',
              background: actionBg,
              border: 'none',
              color: actionInk,
            }}
          >
            {actionLabel}
          </button>
        </div>
      </div>
    </div>
  )
}
