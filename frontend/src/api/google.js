// Google Identity Services (GIS) token-client flow. Loads the GIS script once,
// opens Google's popup, and resolves an OAuth **access token**. That token goes
// to the backend (POST /auth/google), which resolves it to a profile via Google's
// userinfo endpoint. Using the token client (not the ID-token button) is what lets
// us keep a custom-styled button.

const GIS_SRC = 'https://accounts.google.com/gsi/client'
const CLIENT_ID = import.meta.env.VITE_GOOGLE_CLIENT_ID

let gisReady

function loadGis() {
  if (gisReady) return gisReady
  gisReady = new Promise((resolve, reject) => {
    if (window.google?.accounts?.oauth2) return resolve()
    const s = document.createElement('script')
    s.src = GIS_SRC
    s.async = true
    s.defer = true
    s.onload = () => resolve()
    s.onerror = () => reject(new Error('Failed to load Google Identity Services'))
    document.head.appendChild(s)
  })
  return gisReady
}

/** getGoogleToken opens the Google popup and resolves an OAuth access token. */
export async function getGoogleToken() {
  if (!CLIENT_ID) throw new Error('VITE_GOOGLE_CLIENT_ID is not set')
  await loadGis()
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
