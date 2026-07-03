// loadScript injects a <script> once and resolves when it loads. Memoized per
// src, so repeat/concurrent callers share a single load. Shared by the OAuth SDK
// loaders (Google Identity Services, Facebook SDK).

const cache = new Map()

/**
 * @param {string} src
 * @param {{ crossOrigin?: string }} [opts]
 * @returns {Promise<void>} resolves on load, rejects on error
 */
export function loadScript(src, { crossOrigin } = {}) {
  if (cache.has(src)) return cache.get(src)
  const p = new Promise((resolve, reject) => {
    const s = document.createElement('script')
    s.src = src
    s.async = true
    s.defer = true
    if (crossOrigin) s.crossOrigin = crossOrigin
    s.onload = () => resolve()
    s.onerror = () => reject(new Error(`Failed to load script: ${src}`))
    document.head.appendChild(s)
  })
  cache.set(src, p)
  return p
}
