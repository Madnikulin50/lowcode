// Look of the platform. The agent's pages are IFrame blocks of a platform page (the same origin), so the
// branding the platform serves as custom.css (--primary, --black, --light … for the light and the dark mode)
// and the colour mode of the signed-in user (data-color-mode on <html>) are taken from the parent page.
// Opened on its own (fixtures, a direct link) there is no parent: the fallback palette of global.css is
// used and the mode follows prefers-color-scheme.
function parentRoot () {
  try {
    if (window.parent !== window) return window.parent.document.documentElement
  } catch (e) {
    // another origin: the parent is not reachable
  }
  return null
}

function applyMode (mode) {
  const html = document.documentElement
  if (mode) {
    html.setAttribute('data-color-mode', mode)
    html.setAttribute('data-bs-theme', mode)
  } else {
    html.removeAttribute('data-color-mode')
    html.removeAttribute('data-bs-theme')
  }
}

export function installPlatformTheme () {
  const root = parentRoot()
  if (!root) {
    const mq = window.matchMedia?.('(prefers-color-scheme: dark)')
    const follow = () => applyMode(mq?.matches ? 'dark' : 'light')
    follow()
    mq?.addEventListener?.('change', follow)
    return
  }

  // the branding stylesheet is served from a base URL that depends on the deployment: reuse the parent's link
  const link = [...root.ownerDocument.querySelectorAll('link[rel="stylesheet"]')]
    .find(l => /(^|\/)custom\.css(\?|$)/.test(l.getAttribute('href') || ''))
  if (link) {
    const el = document.createElement('link')
    el.rel = 'stylesheet'
    el.href = link.href
    document.head.appendChild(el)
  }

  const sync = () => applyMode(root.getAttribute('data-color-mode') || 'light')
  sync()
  try {
    new MutationObserver(sync).observe(root, { attributes: true, attributeFilter: ['data-color-mode'] })
  } catch (e) {
    // the parent went away: keep what we have
  }
}
