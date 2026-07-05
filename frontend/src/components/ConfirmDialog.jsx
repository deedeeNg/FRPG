import { useEffect, useRef } from 'react'
import { useTheme } from '../theme'

// shadcn AlertDialog look/behavior in the project's inline-style system: modal
// with a dimmed backdrop (no outside-click dismiss), centered card, header
// (title + muted description) and a right-aligned footer (outline Cancel +
// primary Action). Escape triggers Cancel.
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
  const { theme: t } = useTheme()
  const actionRef = useRef(null)

  // Escape cancels; move focus to the action button on open (shadcn behavior).
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

  const actionBg = destructive ? '#dc2626' : t.primary
  const actionHover = destructive ? '#b91c1c' : t.primaryHover

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
        padding: 16,
        fontFamily: "'Public Sans', system-ui, sans-serif",
        background: 'rgba(0,0,0,.5)',
        backdropFilter: 'blur(2px)',
        animation: 'frpgDialogFade .15s ease',
      }}
    >
      <style>{`
        @keyframes frpgDialogFade { from { opacity: 0 } to { opacity: 1 } }
        @keyframes frpgDialogZoom { from { opacity: 0; transform: scale(.96) } to { opacity: 1; transform: scale(1) } }
      `}</style>
      <div
        style={{
          width: '100%',
          maxWidth: 440,
          background: t.surface,
          border: `1px solid ${t.border}`,
          borderRadius: 14,
          boxShadow: t.cardShadow,
          padding: 24,
          animation: 'frpgDialogZoom .15s ease',
        }}
      >
        {/* Header */}
        <h2 id="confirm-title" style={{ margin: 0, fontSize: 18, fontWeight: 700, color: t.ink }}>
          {title}
        </h2>
        {description && (
          <p id="confirm-desc" style={{ margin: '8px 0 0', fontSize: 14, lineHeight: 1.5, color: t.soft }}>
            {description}
          </p>
        )}

        {/* Footer */}
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8, marginTop: 24 }}>
          <button
            type="button"
            onClick={onCancel}
            style={{
              fontFamily: 'inherit',
              fontSize: 14,
              fontWeight: 600,
              padding: '9px 16px',
              borderRadius: 8,
              cursor: 'pointer',
              background: t.surface,
              border: `1px solid ${t.border}`,
              color: t.ink,
            }}
          >
            {cancelLabel}
          </button>
          <button
            ref={actionRef}
            type="button"
            onClick={onConfirm}
            onMouseEnter={(e) => (e.currentTarget.style.background = actionHover)}
            onMouseLeave={(e) => (e.currentTarget.style.background = actionBg)}
            style={{
              fontFamily: 'inherit',
              fontSize: 14,
              fontWeight: 600,
              padding: '9px 16px',
              borderRadius: 8,
              cursor: 'pointer',
              background: actionBg,
              border: '1px solid transparent',
              color: '#fff',
            }}
          >
            {actionLabel}
          </button>
        </div>
      </div>
    </div>
  )
}
