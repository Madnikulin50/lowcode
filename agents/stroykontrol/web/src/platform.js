// The platform user (ТЗ 3.8). The agent's pages are IFrame blocks of a platform page served through
// the API gateway — the same origin — so the token of the signed-in user can be taken from the parent
// application ($auth of its Vue app) and sent to the agent, which checks it and the user's roles.
// Opened on its own (fixtures, a direct link) there is no parent: nothing is sent and the agent,
// when it checks roles, refuses changes.
import { reactive } from 'vue'

export const TOKEN_HEADER = 'X-Platform-Token'

export function platformToken () {
  try {
    if (window.parent === window) return ''
    for (const el of window.parent.document.querySelectorAll('[data-v-app], #app')) {
      const auth = el.__vue_app__?.config?.globalProperties?.$auth
      if (auth?.accessToken) return auth.accessToken
    }
  } catch (e) {
    // another origin: the parent is not reachable
  }
  return ''
}

// installPlatformAuth adds the user's token to every request of the agent's own API.
export function installPlatformAuth () {
  const orig = window.fetch.bind(window)
  window.fetch = (input, init = {}) => {
    const url = typeof input === 'string' ? input : input?.url || ''
    if (/^(\.\/)?api\//.test(url) || url.startsWith(new URL('api/', location.href).href)) {
      const tok = platformToken()
      if (tok) {
        const headers = new Headers(init.headers || (typeof input !== 'string' ? input.headers : undefined))
        headers.set(TOKEN_HEADER, tok)
        init = { ...init, headers }
      }
    }
    return orig(input, init)
  }
}

// me: who is working and what they may do; checked=false — the agent does not check roles (all allowed).
export const me = reactive({ loaded: false, checked: false, label: '', roleLabels: [], allowed: [], error: '' })

export async function loadMe () {
  try {
    const r = await fetch('api/me')
    const d = await r.json().catch(() => ({}))
    if (!r.ok) throw new Error(d.error || r.statusText)
    Object.assign(me, { loaded: true, checked: !!d.checked, label: d.label || '', roleLabels: d.roleLabels || [], allowed: d.allowed || [], error: '' })
  } catch (e) {
    Object.assign(me, { loaded: true, checked: true, allowed: [], error: e.message })
  }
}

export const can = (act) => !me.checked || me.allowed.includes(act)
