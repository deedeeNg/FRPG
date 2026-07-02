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
