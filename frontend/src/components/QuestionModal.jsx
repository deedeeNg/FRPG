import { useState } from 'react'
import { useLanguage } from '../i18n'
import { hudColors, roundCorners, glassDark, glassTextShadow } from '../hud'
import { gradeExercise } from '../api/exercise'
import ClockFace from './ClockFace'

// Centered modal overlay showing one exercise — multiple_choice (pick a choice)
// or fill_blank (type the answer), whichever the served item's format is. A
// content.clock payload (telling_time) renders an SVG clock face above the
// prompt. On submit it grades server-side, shows the result, then reports it up
// via onResult(correct) — the caller (Map) advances on true and resets on false.
// UI-only beyond the grade call.
//   exercise: the served item (answer stripped) | null while loading
//   loading, error: fetch state from the caller
//   onResult(correct: boolean), onClose()
export default function QuestionModal({ exercise, loading, error, onResult, onClose }) {
  const { t: tr } = useLanguage()
  const [selected, setSelected] = useState(null)
  const [typed, setTyped] = useState('')
  const [result, setResult] = useState(null) // null | { correct }
  const [submitting, setSubmitting] = useState(false)

  const isFillBlank = exercise?.format === 'fill_blank'
  const choices = exercise?.content?.choices ?? []
  const clock = exercise?.content?.clock
  const canSubmit = isFillBlank ? typed.trim() !== '' : selected !== null

  const submit = async () => {
    if (!canSubmit || submitting) return
    setSubmitting(true)
    try {
      const answer = isFillBlank ? { text: typed } : { selected: [selected] }
      const correct = await gradeExercise(exercise.exerciseId, answer)
      setResult({ correct })
    } catch {
      setResult({ correct: false })
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div
      onClick={onClose}
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 50,
        background: 'rgba(10,8,20,0.55)',
        backdropFilter: 'blur(2px)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 20,
      }}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        style={{
          ...glassDark,
          ...roundCorners,
          width: '100%',
          maxWidth: 460,
          padding: 24,
          fontFamily: "'Pixelify Sans', sans-serif",
          color: '#ffffff',
        }}
      >
        {loading || !exercise ? (
          <p style={{ margin: 0, textAlign: 'center', textShadow: glassTextShadow }}>
            {error ? tr('map.q.error') : tr('map.q.loading')}
          </p>
        ) : result ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 18, alignItems: 'center' }}>
            <div
              style={{
                fontSize: 18,
                fontWeight: 700,
                textAlign: 'center',
                textShadow: glassTextShadow,
                color: result.correct ? hudColors.green : hudColors.rose,
              }}
            >
              {tr(result.correct ? 'map.q.correct' : 'map.q.wrong')}
            </div>
            <button type="button" onClick={() => onResult(result.correct)} style={primaryBtn}>
              {tr('map.q.continue')}
            </button>
          </div>
        ) : (
          <>
            <div style={{ fontSize: 14, color: hudColors.gold, fontWeight: 700, textShadow: glassTextShadow, marginBottom: 6 }}>
              {exercise.prompt.instructions}
            </div>

            {clock && (
              <div style={{ display: 'flex', justifyContent: 'center', margin: '4px 0 18px' }}>
                <ClockFace hour={clock.hour} minute={clock.minute} />
              </div>
            )}

            {exercise.prompt.text && (
              <div style={{ fontSize: 22, fontWeight: 700, marginBottom: 18, textShadow: glassTextShadow }}>
                {exercise.prompt.text}
              </div>
            )}

            {isFillBlank ? (
              <>
                <div style={{ fontSize: 22, fontWeight: 700, marginBottom: 18, textShadow: glassTextShadow }}>
                  {exercise.content.template}
                </div>
                <input
                  type="text"
                  value={typed}
                  onChange={(e) => setTyped(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && submit()}
                  placeholder={tr('map.q.typeAnswer')}
                  autoFocus
                  style={inputStyle}
                />
              </>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 10, marginBottom: 20 }}>
                {choices.map((c) => {
                  const active = selected === c.id
                  return (
                    <button
                      key={c.id}
                      type="button"
                      onClick={() => setSelected(c.id)}
                      style={{
                        ...choiceBtn,
                        background: active ? 'rgba(244,197,66,0.28)' : 'rgba(255,255,255,0.08)',
                        border: `1.5px solid ${active ? hudColors.gold : 'rgba(255,255,255,0.3)'}`,
                      }}
                    >
                      {c.text}
                    </button>
                  )
                })}
              </div>
            )}

            <button
              type="button"
              onClick={submit}
              disabled={!canSubmit || submitting}
              style={{ ...primaryBtn, opacity: !canSubmit || submitting ? 0.5 : 1, cursor: !canSubmit || submitting ? 'default' : 'pointer' }}
            >
              {tr('map.q.submit')}
            </button>
          </>
        )}
      </div>
    </div>
  )
}

const choiceBtn = {
  fontFamily: 'inherit',
  fontSize: 16,
  color: '#ffffff',
  textShadow: glassTextShadow,
  padding: '12px 16px',
  borderRadius: 10,
  textAlign: 'left',
  cursor: 'pointer',
  transition: 'background .08s',
}

const inputStyle = {
  fontFamily: 'inherit',
  fontSize: 16,
  color: '#ffffff',
  background: 'rgba(255,255,255,0.08)',
  border: `1.5px solid ${hudColors.gold}`,
  borderRadius: 10,
  padding: '12px 16px',
  marginBottom: 20,
  width: '100%',
  boxSizing: 'border-box',
  outline: 'none',
}

const primaryBtn = {
  fontFamily: 'inherit',
  fontSize: 15,
  fontWeight: 700,
  color: hudColors.ink,
  background: hudColors.gold,
  border: 'none',
  borderRadius: 10,
  padding: '11px 22px',
  cursor: 'pointer',
  width: '100%',
}
