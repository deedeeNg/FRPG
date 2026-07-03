// Talks to the Go backend's POST /auth/{provider} endpoint. In dev these are
// same-origin thanks to the vite proxy (/auth -> backend), so no CORS is needed.

/**
 * login sends credentials to a provider and returns the session token.
 * @param {'local'|'google'|'facebook'} provider
 * @param {object} credentials  e.g. { email, password } for local, { token } for social
 * @returns {Promise<string>} the session JWT
 * @throws {Error} with the backend's error message on non-2xx
 */
export async function login(provider, credentials) {
  const res = await fetch(`/auth/${provider}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(credentials),
  })

  let data = {}
  try {
    data = await res.json()
  } catch {
    // response had no/invalid JSON body
  }

  if (!res.ok) {
    throw new Error(data.error || `Login failed (${res.status})`)
  }
  return data.token
}

/**
 * signup creates a local (email + password) account and returns a session token
 * (the backend logs the new user straight in).
 * @param {string} email
 * @param {string} password
 * @returns {Promise<string>} the session JWT
 * @throws {Error} with the backend's error message on non-2xx (e.g. email taken)
 */
export async function signup(email, password) {
  const res = await fetch('/signup', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })

  let data = {}
  try {
    data = await res.json()
  } catch {
    // response had no/invalid JSON body
  }

  if (!res.ok) {
    throw new Error(data.error || `Sign-up failed (${res.status})`)
  }
  return data.token
}

/**
 * fetchMe verifies the stored session token against the backend (GET /api/me).
 * Returns the user ({ userId, email }) when the session is valid, or null when
 * there's no token or it's rejected — clearing a stale token on the way out.
 * Never throws.
 * @returns {Promise<{userId: string, email: string} | null>}
 */
export async function fetchMe() {
  const token = localStorage.getItem('frpg_token')
  if (!token) return null
  try {
    const res = await fetch('/api/me', { headers: { Authorization: `Bearer ${token}` } })
    if (!res.ok) {
      localStorage.removeItem('frpg_token')
      return null
    }
    return await res.json()
  } catch {
    return null
  }
}
