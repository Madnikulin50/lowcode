import { apiClients } from 'corteza-lib/js/dist'

let _ComposeAPI = null
let _SystemAPI = null

function getAccessToken() {
  const auth = window.__auth
  return auth?.accessTokenFn?.()
}

export function getComposeAPI () {
  if (!_ComposeAPI) {
    _ComposeAPI = new apiClients.Compose({
      baseURL: `${window.CortezaAPI}/compose`,
      accessTokenFn: getAccessToken,
    })
  }
  return _ComposeAPI
}

export function getSystemAPI () {
  if (!_SystemAPI) {
    _SystemAPI = new apiClients.System({
      baseURL: `${window.CortezaAPI}/system`,
      accessTokenFn: getAccessToken,
    })
  }
  return _SystemAPI
}

export function setComposeAPI (api) {
  _ComposeAPI = api
}

export function setSystemAPI (api) {
  _SystemAPI = api
}
