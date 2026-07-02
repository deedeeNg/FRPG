// Google Identity Services (GIS) token-client flow. Loads the GIS script once,
// opens Google's popup, and resolves an OAuth **access token**. That token goes
// to the backend (POST /auth/google), which resolves it to a profile via Google's
// userinfo endpoint. Using the token client (not the ID-token button) is what lets
// us keep a custom-styled button.

import { loadScript } from './loadScript'

const GIS_SRC = 'https://accounts.google.com/gsi/client'
const CLIENT_ID = import.meta.env.VITE_GOOGLE_CLIENT_ID

/** getGoogleToken opens the Google popup and resolves an OAuth access token. */
export async function getGoogleToken() {
  if (!CLIENT_ID) throw new Error('VITE_GOOGLE_CLIENT_ID is not set')
  if (!window.google?.accounts?.oauth2) await loadScript(GIS_SRC)
  return new Promise((resolve, reject) => {
    const client = window.google.accounts.oauth2.initTokenClient({
      client_id: CLIENT_ID,
      scope: 'openid email profile',
      callback: (resp) => {
        if (resp.error) return reject(new Error(resp.error))
        resolve(resp.access_token)
      },
      error_callback: (err) => reject(new Error(err?.message || 'Google sign-in cancelled')),
    })
    client.requestAccessToken()
  })
}
