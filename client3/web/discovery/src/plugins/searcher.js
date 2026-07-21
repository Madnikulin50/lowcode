import axios from 'axios'

export default function (opt = {}) {
  return {
    install(app, options) {
      if (!opt.baseURL) {
        if (!window.CortezaDiscoveryAPI) throw new Error('window.CortezaDiscoveryAPI not set')
        opt.baseURL = `${window.CortezaDiscoveryAPI}/`
      }
      if (!opt.accessTokenFn && options?.accessTokenFn) {
        opt.accessTokenFn = options.accessTokenFn
      }
      app.config.globalProperties.$DiscoveryAPI = new Searcher(opt)
      window.__discoveryAPI = app.config.globalProperties.$DiscoveryAPI
    }
  }
}

function stdResolve(response) {
  if (response.data.error) {
    return Promise.reject(response.data.error)
  } else {
    return response.data.response
  }
}

class Searcher {
  baseURL
  accessTokenFn
  headers = {}

  constructor({ baseURL, headers, accessTokenFn }) {
    this.baseURL = baseURL
    this.accessTokenFn = accessTokenFn
    this.headers = {
      'Content-Type': 'application/json',
    }
    this.setHeaders(headers)
  }

  setAccessTokenFn(fn) {
    this.accessTokenFn = fn
    return this
  }

  setHeaders(headers) {
    if (typeof headers === 'object') {
      this.headers = headers
    }
    return this
  }

  setHeader(name, value) {
    if (value === undefined) {
      delete this.headers[name]
    } else {
      this.headers[name] = value
    }
    return this
  }

  api() {
    const headers = { ...this.headers }
    const accessToken = this.accessTokenFn ? this.accessTokenFn() : undefined
    if (accessToken) {
      headers.Authorization = 'Bearer ' + accessToken
    }
    return axios.create({
      withCredentials: true,
      baseURL: this.baseURL,
      headers,
    })
  }

  async query(a, extra = {}) {
    const {
      modules,
      namespaces,
      from,
      size,
      resourceTypes,
    } = a || {}

    const params = new URLSearchParams()

    if (modules?.length > 0) modules.forEach(m => params.append('moduleAggs', m))
    if (namespaces?.length > 0) namespaces.forEach(n => params.append('namespaceAggs', n))
    if (resourceTypes?.length > 0) resourceTypes.forEach(t => params.append('resourceTypes', t))

    if (from) params.append('from', from)
    if (size) params.append('size', size)

    const cfg = {
      ...extra,
      method: 'get',
      url: this.queryEndpoint(a),
      params,
    }

    return this.api().request(cfg).then(result => stdResolve(result))
  }

  queryEndpoint(a) {
    const { query = '' } = a || {}
    return `/?q=${query}`
  }
}