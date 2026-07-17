// Persists adventure-map progress to localStorage so logging out and back in
// restores both the current node and the remaining attempts. The map layout is
// deterministic (fixed seed in data/map.js), so node ids like "3-1" are stable
// across sessions — safe to persist by id.
//
// Shape: { current: "3-1", attempts: 2 }
//   current  — id of the "you are here" node.
//   attempts — remaining tries in the single global pool for the whole map
//              (not per-node). Starts at MAX_ATTEMPTS; a wrong answer decrements
//              it, and hitting 0 resets the map (see Map.jsx resolveAttempt).
//
// Keyed per user (frpg_map_<userId>) so two accounts on the same browser keep
// separate progress. Not cleared on logout — the entry survives so re-login
// restores where the player left off.

// Single global attempt pool for the whole map.
export const MAX_ATTEMPTS = 3

const keyFor = (userId) => `frpg_map_${userId}`

// Reads stored progress for a user, or null when nothing is stored / unparseable.
export function read(userId) {
  try {
    const raw = localStorage.getItem(keyFor(userId))
    return raw ? JSON.parse(raw) : null
  } catch {
    return null
  }
}

// Writes progress for a user. state = { current, attempts }.
export function write(userId, state) {
  try {
    localStorage.setItem(keyFor(userId), JSON.stringify(state))
  } catch {
    // Storage full / unavailable (private mode): progress just won't persist.
  }
}
