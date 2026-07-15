// Procedural adventure-map graph: columns run left→right, each column holding a
// few branching nodes wired to the next column — the same node-map interaction
// as a Slay-the-Spire-style run map, rotated 90° ("sideways": progression reads
// left to right instead of bottom to top). Pure data; Map.jsx lays it out.
//
// Each node is typed by *skill* (reading/listening/speaking/writing) — the same
// four skills as the Home quest cards and the backend's Exercise.Skill — so a
// node represents one exercise of that type, not a dungeon-crawler encounter.
//
// This is placeholder demo data (like data/skills.js, data/stats.js, data/user.js)
// — a real run map would come from the backend once quest content exists.

// Small seeded PRNG so the generated layout is stable across re-renders instead
// of reshuffling on every render.
function mulberry32(seed) {
  let s = seed
  return function rng() {
    s |= 0
    s = (s + 0x6d2b79f5) | 0
    let t = Math.imul(s ^ (s >>> 15), 1 | s)
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

// The four question/skill types a node can be — same keys as data/skills.js
// (speaking/listening/reading/writing), so labels and colors stay in sync with
// the rest of the app via `skill.<key>` i18n strings.
export const SKILL_TYPES = ['reading', 'listening', 'speaking', 'writing']

function pickType(rng) {
  return SKILL_TYPES[Math.floor(rng() * SKILL_TYPES.length)]
}

const clamp = (n, lo, hi) => Math.max(lo, Math.min(hi, n))

/**
 * Generates a branching node graph, columns left to right.
 *   columns — number of stages (the horizontal/"sideways" progression)
 *   seed — deterministic PRNG seed; same seed always produces the same layout
 * Returns { nodes: [{id,col,row,rows,type,jitter}], edges: [{from,to}], columns }
 */
export function generateMap({ columns = 8, seed = 20260716 } = {}) {
  const rng = mulberry32(seed)
  const colRows = Array.from({ length: columns }, (_, c) => (c === 0 ? 1 : 2 + Math.floor(rng() * 3)))

  const nodes = []
  colRows.forEach((rows, col) => {
    for (let row = 0; row < rows; row++) {
      nodes.push({
        id: `${col}-${row}`,
        col,
        row,
        rows,
        type: pickType(rng),
        // Small vertical offset per node so paths meander rather than forming a
        // rigid grid (purely cosmetic, generated once with the same rng stream).
        jitter: (rng() - 0.5) * 18,
      })
    }
  })

  const edges = []
  for (let col = 0; col < columns - 1; col++) {
    const from = nodes.filter((n) => n.col === col)
    const to = nodes.filter((n) => n.col === col + 1)

    from.forEach((n) => {
      const rel = n.row / Math.max(1, n.rows - 1)
      const target = clamp(Math.round(rel * (to.length - 1)), 0, to.length - 1)
      const targets = new Set([target])
      if (to.length > 1 && rng() < 0.45) {
        targets.add(clamp(target + (rng() < 0.5 ? -1 : 1), 0, to.length - 1))
      }
      targets.forEach((t) => edges.push({ from: n.id, to: to[t].id }))
    })

    // Every node in the next column needs at least one incoming path, so nothing
    // is unreachable.
    to.forEach((n, i) => {
      if (edges.some((e) => e.to === n.id)) return
      const rel = i / Math.max(1, to.length - 1)
      const source = from[clamp(Math.round(rel * (from.length - 1)), 0, from.length - 1)]
      edges.push({ from: source.id, to: n.id })
    })
  }

  return { nodes, edges, columns }
}
