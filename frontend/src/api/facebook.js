// Facebook JS SDK login flow. Loads the SDK once, opens Facebook's popup with the
// `email` scope, and resolves an **access token**. That token goes to the backend
// (POST /auth/facebook), which resolves it to a profile via the Graph /me endpoint.
//
// Needs a Facebook **App ID** (public, VITE_FACEBOOK_APP_ID) — the App Secret is
// never used in the browser. The app must have Facebook Login + the `email`
// permission enabled, or the token carries no email and the backend rejects it.

import { loadScript } from './loadScript'

const FB_SRC = 'https://connect.facebook.net/en_US/sdk.js'
const APP_ID = import.meta.env.VITE_FACEBOOK_APP_ID

let fbReady

// The FB SDK signals readiness via window.fbAsyncInit (where we must call
// FB.init), not the script's onload — so we resolve there and use loadScript only
// to inject the tag and surface load errors.
function loadFb() {
  if (fbReady) return fbReady
  fbReady = new Promise((resolve, reject) => {
    if (window.FB) return resolve()
    window.fbAsyncInit = () => {
      window.FB.init({ appId: APP_ID, cookie: true, xfbml: false, version: 'v21.0' })
      resolve()
    }
    loadScript(FB_SRC, { crossOrigin: 'anonymous' }).catch(reject)
  })
  return fbReady
}

/** getFacebookToken opens the Facebook popup and resolves an access token. */
export async function getFacebookToken() {
  if (!APP_ID) throw new Error('VITE_FACEBOOK_APP_ID is not set')
  await loadFb()
  return new Promise((resolve, reject) => {
    window.FB.login(
      (resp) => {
        if (resp.authResponse?.accessToken) {
          resolve(resp.authResponse.accessToken)
        } else {
          reject(new Error('Facebook sign-in cancelled'))
        }
      },
      { scope: 'email' },
    )
  })
}
