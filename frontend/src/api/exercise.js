// Talks to the backend's exercise endpoints (proxied same-origin via vite:
// /api -> backend). The answer key is never sent to the client — grading is a
// server round-trip.

function authHeaders(extra = {}) {
  const token = localStorage.getItem('frpg_token')
  return token ? { ...extra, Authorization: `Bearer ${token}` } : extra
}

/**
 * fetchExercise loads one exercise to attempt (answer stripped by the server).
 * @returns {Promise<object>} the exercise (exerciseId, prompt, content, …)
 * @throws {Error} on non-2xx
 */
export async function fetchExercise() {
  const res = await fetch('/api/exercise', { headers: authHeaders() })
  if (!res.ok) throw new Error(`Could not load a question (${res.status})`)
  return res.json()
}

/**
 * gradeExercise submits an answer and returns whether it was correct. Pass
 * `{ selected }` (chosen choice ids) for multiple_choice, or `{ text }` (the
 * typed answer) for fill_blank — whichever matches the exercise's own format.
 * @param {string} exerciseId
 * @param {{selected?: string[], text?: string}} answer
 * @returns {Promise<boolean>} true when correct
 * @throws {Error} on non-2xx
 */
export async function gradeExercise(exerciseId, answer) {
  const res = await fetch('/api/exercise/grade', {
    method: 'POST',
    headers: authHeaders({ 'Content-Type': 'application/json' }),
    body: JSON.stringify({ exerciseId, ...answer }),
  })
  if (!res.ok) throw new Error(`Could not grade (${res.status})`)
  const data = await res.json()
  return data.correct === true
}
