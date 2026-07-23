import { useEffect, useMemo, useRef, useState } from 'react'
import { useLanguage } from '../i18n'
import HudLayout from '../components/HudLayout'
import { hudColors, roundCorners, liquidGlass, glassTextShadow, pixelated } from '../hud'
import { useWindowWidth, HUD_BREAKPOINT } from '../hooks/useWindowWidth'
import { generateMap, SKILL_TYPES } from '../data/map'
import { read as readProgress, write as writeProgress, MAX_ATTEMPTS } from '../data/mapProgress'
import QuestionModal from '../components/QuestionModal'
import ConfirmDialog from '../components/ConfirmDialog'
import { fetchExercise } from '../api/exercise'
import iconSpeech from '../assets/hud/icon-speech.png'
import iconNote from '../assets/hud/icon-note.png'
import iconBook from '../assets/hud/icon-book.png'
import iconQuill from '../assets/hud/icon-quill.png'

// Pixel-art heart bitmap (8×7): 0 = empty, 1 = outline, 2 = fill. Drawn as unit
// rects in an 8×7 viewBox so it scales crisply like the rest of the pixel art.
const HEART_BITMAP = [
  '01100110',
  '12211221',
  '12222221',
  '12222221',
  '01222210',
  '00122100',
  '00011000',
]

// One life heart. `filled` = a remaining attempt (red); otherwise a spent slot
// (grey), matching the classic "❤❤♡" life bar in the reference.
function Heart({ filled, cell = 5 }) {
  const outline = 'rgba(43,36,64,0.85)'
  const fill = filled ? '#e0384f' : 'rgba(220,220,224,0.85)'
  const w = 8
  const h = 7
  return (
    <svg
      width={w * cell}
      height={h * cell}
      viewBox={`0 0 ${w} ${h}`}
      shapeRendering="crispEdges"
      style={{ ...pixelated, filter: 'drop-shadow(0 1px 1px rgba(43,36,64,0.55))' }}
      aria-hidden="true"
    >
      {HEART_BITMAP.flatMap((r, y) =>
        [...r].map((c, x) =>
          c === '0' ? null : (
            <rect key={`${x}-${y}`} x={x} y={y} width="1" height="1" fill={c === '1' ? outline : fill} />
          )
        )
      )}
    </svg>
  )
}

// Same icon + chip color per skill as the Home quest cards, so a map node reads
// as "an exercise of this type" using the app's existing visual language.
const SKILL_STYLE = {
  speaking: { icon: iconSpeech, color: hudColors.blue },
  listening: { icon: iconNote, color: hudColors.gold, filter: 'brightness(0.2)' },
  reading: { icon: iconBook, color: hudColors.green },
  writing: { icon: iconQuill, color: hudColors.rose },
}

// Virtual coordinate space for the node graph (px). Columns run left→right —
// the sideways read of a Slay-the-Spire-style run map (reference: a map of the
// same shape, but progressing bottom→top). The whole graph is then scaled as one
// unit to fit its frame (see the fit-scale below), so these are just the intrinsic
// proportions, not final on-screen sizes.
const COL_GAP = 140
const ROW_GAP = 96
const NODE = 46
const PAD = 40

function layout(map) {
  const maxRows = Math.max(...map.nodes.map((n) => n.rows))
  const centerY = PAD + ((maxRows - 1) * ROW_GAP) / 2
  const points = {}
  map.nodes.forEach((n) => {
    points[n.id] = {
      x: PAD + n.col * COL_GAP,
      y: centerY + (n.row - (n.rows - 1) / 2) * ROW_GAP + n.jitter,
    }
  })
  return {
    points,
    width: PAD * 2 + (map.columns - 1) * COL_GAP,
    height: PAD * 2 + (maxRows - 1) * ROW_GAP,
  }
}

// Measures an element's content box and keeps it in state, so the graph can be
// scaled to fit whatever size the frame ends up (fully flexible — no fixed sizes,
// re-fits on any resize).
function useContentSize(ref) {
  const [size, setSize] = useState({ w: 0, h: 0 })
  useEffect(() => {
    const el = ref.current
    if (!el) return
    const ro = new ResizeObserver(([entry]) => {
      const { width, height } = entry.contentRect
      setSize({ w: width, h: height })
    })
    ro.observe(el)
    return () => ro.disconnect()
  }, [ref])
  return size
}

/**
 * Adventure map — a branching node graph the player advances through left to
 * right, each node typed by skill (reading / listening / speaking / writing — the
 * same four as the Home quest cards). The whole graph is scaled as a single unit
 * to fit its frame and centered both axes, so it fills the box and stays centered
 * at any viewport size (fit-scale via ResizeObserver; no hardcoded dimensions).
 * The frame occupies ~80% of the available box. UI-only: node choice is local
 * state, not wired to any quest/backend logic yet.
 *   activeRoute, onNavigate(route), onLogout — HUD nav (same contract as Home)
 *   userId — persists progress per user in localStorage (see data/mapProgress.js)
 */
export default function Map({ activeRoute = 'map', onNavigate, onLogout, userId }) {
  const { t: tr } = useLanguage()
  const compact = useWindowWidth() < HUD_BREAKPOINT

  const map = useMemo(() => generateMap({ columns: 8 }), [])
  const { points, width, height } = useMemo(() => layout(map), [map])

  const cardRef = useRef(null)
  const { w: boxW, h: boxH } = useContentSize(cardRef)

  // Fit the whole graph inside the frame, centered. Cap the upscale so nodes
  // never balloon on very wide screens; otherwise scale freely so it always fits.
  const scale = useMemo(() => {
    if (!boxW || !boxH) return 1
    return Math.min(1.7, boxW / width, boxH / height)
  }, [boxW, boxH, width, height])

  const startId = map.nodes[0].id
  // Restore saved progress on mount (per user), falling back to a fresh run.
  const saved = useMemo(() => readProgress(userId), [userId])
  const [current, setCurrent] = useState(saved?.current ?? startId) // "you are here"
  const [attempts, setAttempts] = useState(saved?.attempts ?? MAX_ATTEMPTS) // global pool
  const [hovered, setHovered] = useState(null)
  const [confirmRestart, setConfirmRestart] = useState(false)

  // Persist progress whenever the current node or the attempt pool changes.
  useEffect(() => {
    writeProgress(userId, { current, attempts })
  }, [userId, current, attempts])

  // Question modal state: the node being attempted, plus the fetched exercise.
  const [attempt, setAttempt] = useState(null)
  const [exercise, setExercise] = useState(null)
  const [qLoading, setQLoading] = useState(false)
  const [qError, setQError] = useState(false)

  // Only nodes one step ahead of the current position are selectable — the same
  // "next step lights up" read as the reference map.
  const reachable = useMemo(
    () => new Set(map.edges.filter((e) => e.from === current).map((e) => e.to)),
    [map, current]
  )

  // Clicking a reachable node opens the question modal (fetches one exercise).
  const attemptNode = async (nodeId) => {
    setAttempt(nodeId)
    setExercise(null)
    setQError(false)
    setQLoading(true)
    try {
      setExercise(await fetchExercise())
    } catch {
      setQError(true)
    } finally {
      setQLoading(false)
    }
  }

  const closeModal = () => {
    setAttempt(null)
    setExercise(null)
    setQError(false)
  }

  // A node in the last column completes the run.
  const isFinalNode = (id) => map.nodes.find((n) => n.id === id)?.col === map.columns - 1

  // Answered. Correct: advance to the attempted node (pool unchanged) — unless it
  // was the final node, which finishes the run and resets the whole map (back to
  // start, pool refilled). Wrong: decrement the shared pool; when it hits 0 the
  // map resets the same way.
  const resolveAttempt = (correct) => {
    if (correct && !isFinalNode(attempt)) {
      setCurrent(attempt)
    } else if (correct || attempts <= 1) {
      setCurrent(startId)
      setAttempts(MAX_ATTEMPTS)
    } else {
      setAttempts(attempts - 1)
    }
    closeModal()
  }

  // Manual restart: same reset path as finishing/zero-attempts above, but
  // player-triggered and guarded by a confirm prompt. Persisted via the effect.
  const restart = () => {
    setCurrent(startId)
    setAttempts(MAX_ATTEMPTS)
    setConfirmRestart(false)
  }

  const boxStyle = compact
    ? { padding: '0 20px 28px', boxSizing: 'border-box' }
    : { position: 'absolute', left: 20, right: 20, top: 116, bottom: 20 }

  // Legend sizes scale down on compact viewports so the overlay stays a small
  // corner chip instead of swallowing the (proportionally smaller) map frame.
  const lg = compact
    ? { pad: '7px 9px', gap: 4, minW: undefined, title: 10, item: 10.5, chip: 14, img: 8 }
    : { pad: '12px 16px', gap: 8, minW: 148, title: 12, item: 12.5, chip: 18, img: 10 }

  // Pinned to the top-right of the whole map area (mirrors the life bar top-left),
  // sitting just below the HUD nav bar — not inside the map frame.
  const legend = (
    <div
      style={{
        ...liquidGlass,
        ...roundCorners,
        position: 'absolute',
        top: 12,
        right: 12,
        maxWidth: 'calc(100% - 24px)',
        padding: lg.pad,
        display: 'flex',
        flexDirection: 'column',
        gap: lg.gap,
        minWidth: lg.minW,
      }}
    >
      <div
        style={{
          fontSize: lg.title,
          fontWeight: 700,
          color: hudColors.gold,
          letterSpacing: 1,
          textShadow: glassTextShadow,
          marginBottom: 2,
        }}
      >
        {tr('map.legend.title')}
      </div>
      {SKILL_TYPES.map((key) => {
        const { icon, color, filter } = SKILL_STYLE[key]
        return (
          <div key={key} style={{ display: 'flex', alignItems: 'center', gap: lg.gap, fontSize: lg.item, color: '#ffffff', textShadow: glassTextShadow }}>
            <span style={{ width: lg.chip, height: lg.chip, flex: 'none', background: color, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <img src={icon} alt="" style={{ ...pixelated, width: lg.img, height: lg.img, filter }} />
            </span>
            {tr('skill.' + key)}
          </div>
        )
      })}
    </div>
  )

  // Life bar: one heart per attempt in the global pool, filled = remaining.
  // Top-left, just below the HUD navbar (mirrors the legend on the right).
  const lives = (
    <div
      role="img"
      aria-label={`${attempts}/${MAX_ATTEMPTS} ${tr('map.attempts')}`}
      style={{ position: 'absolute', top: 12, left: 12, display: 'flex', gap: 6 }}
    >
      {Array.from({ length: MAX_ATTEMPTS }).map((_, i) => (
        <Heart key={i} filled={i < attempts} />
      ))}
    </div>
  )

  const returnBtn = (
    <button
      type="button"
      onClick={() => onNavigate('home')}
      style={{
        ...liquidGlass,
        position: 'absolute',
        bottom: 12,
        left: 12,
        fontFamily: 'inherit',
        fontSize: 13.5,
        fontWeight: 700,
        color: '#ffffff',
        textShadow: glassTextShadow,
        border: `1px solid ${hudColors.gold}`,
        borderRadius: 10,
        padding: '9px 18px',
        cursor: 'pointer',
      }}
    >
      ← {tr('map.return')}
    </button>
  )

  // Player-triggered restart. Mirrors the return button (bottom-right) with a
  // rose accent to signal it's destructive; opens a confirm prompt before reset.
  const restartBtn = (
    <button
      type="button"
      onClick={() => setConfirmRestart(true)}
      style={{
        ...liquidGlass,
        position: 'absolute',
        bottom: 12,
        right: 12,
        fontFamily: 'inherit',
        fontSize: 13.5,
        fontWeight: 700,
        color: '#ffffff',
        textShadow: glassTextShadow,
        border: `1px solid ${hudColors.rose}`,
        borderRadius: 10,
        padding: '9px 18px',
        cursor: 'pointer',
      }}
    >
      ↻ {tr('map.restart')}
    </button>
  )

  // The graph at its intrinsic size; scaled + centered as one unit by the card.
  const graph = (
    <div style={{ position: 'relative', width, height, flex: 'none', transform: `scale(${scale})`, transformOrigin: 'center center' }}>
      <svg width={width} height={height} style={{ position: 'absolute', left: 0, top: 0, overflow: 'visible' }}>
        {map.edges.map((e) => {
          const a = points[e.from]
          const b = points[e.to]
          const active = e.from === current
          return (
            <line
              key={`${e.from}-${e.to}`}
              x1={a.x}
              y1={a.y}
              x2={b.x}
              y2={b.y}
              stroke={active ? hudColors.gold : 'rgba(255,255,255,0.3)'}
              strokeWidth={active ? 2.5 : 1.5}
              strokeDasharray="5 5"
            />
          )
        })}
      </svg>

      {map.nodes.map((n) => {
        const p = points[n.id]
        const { icon, color, filter } = SKILL_STYLE[n.type]
        const isCurrent = n.id === current
        const isReachable = reachable.has(n.id)
        const isHover = hovered === n.id
        const label = tr('skill.' + n.type)
        return (
          <div
            key={n.id}
            role="button"
            tabIndex={0}
            title={label}
            onMouseEnter={() => setHovered(n.id)}
            onMouseLeave={() => setHovered(null)}
            onClick={() => isReachable && attemptNode(n.id)}
            onKeyDown={(e) => (e.key === 'Enter' || e.key === ' ') && isReachable && attemptNode(n.id)}
            style={{
              position: 'absolute',
              left: p.x - NODE / 2,
              top: p.y - NODE / 2,
              width: NODE,
              height: NODE,
              borderRadius: '50%',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              background: isCurrent ? color : isReachable ? 'rgba(255,255,255,0.16)' : 'rgba(28,22,52,0.4)',
              border: `2px solid ${isCurrent ? hudColors.gold : isReachable ? 'rgba(255,255,255,0.7)' : 'rgba(255,255,255,0.25)'}`,
              opacity: isReachable || isCurrent ? 1 : 0.55,
              cursor: isReachable ? 'pointer' : 'default',
              boxShadow: isHover && isReachable ? '0 0 0 4px rgba(244,197,66,0.25)' : 'none',
              transition: 'background .1s, box-shadow .1s',
            }}
          >
            <img src={icon} alt="" style={{ ...pixelated, width: 20, height: 20, filter }} />
          </div>
        )
      })}
    </div>
  )

  return (
    <HudLayout activeRoute={activeRoute} onNavigate={onNavigate} onLogout={onLogout}>
      <div style={boxStyle}>
        {/* "The box": the full available area. The frame below is centered inside
            it at ~80% of its size; the graph is then fit-scaled + centered within
            the frame, so the map fills the box and stays centered on any screen. */}
        <div style={{ position: 'relative', width: '100%', height: compact ? '72vh' : '100%', minHeight: compact ? 440 : undefined }}>
          <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <div style={{ position: 'relative', width: '80%', height: '80%', minWidth: 280, minHeight: 240 }}>
              <div
                ref={cardRef}
                style={{
                  ...liquidGlass,
                  ...roundCorners,
                  width: '100%',
                  height: '100%',
                  overflow: 'hidden',
                  boxSizing: 'border-box',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                {graph}
              </div>
            </div>
          </div>
          {lives}
          {legend}
          {returnBtn}
          {restartBtn}
        </div>
      </div>

      {attempt && (
        <QuestionModal
          exercise={exercise}
          loading={qLoading}
          error={qError}
          attempts={attempts}
          final={isFinalNode(attempt)}
          onResult={resolveAttempt}
          onClose={closeModal}
        />
      )}

      <ConfirmDialog
        open={confirmRestart}
        title={tr('map.restart.title')}
        description={tr('map.restart.desc')}
        cancelLabel={tr('common.cancel')}
        actionLabel={tr('map.restart.action')}
        destructive
        onCancel={() => setConfirmRestart(false)}
        onConfirm={restart}
      />
    </HudLayout>
  )
}
