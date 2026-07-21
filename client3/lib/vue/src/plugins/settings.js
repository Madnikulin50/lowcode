import { reactive } from 'vue'

const localAttachment = /^attachment:(\d+)/

export class Settings {
  constructor () {
    this.current = {}
    this.api = undefined
    return this
  }

  async init ({ api }) {
    this.api = api
    if (!this.api) {
      throw new Error('api.notDefined')
    }

    return this.fetch()
  }

  async fetch () {
    return this.api.settingsCurrent()
      .then(settings => {
        this.current = settings || {}
        return settings
      })
  }

  get (k, d) {
    k = k.split(/\./g)
    let s = this.current

    for (let i = 0; i < k.length - 1; i++) {
      const p = k[i]
      s = s[p] || {}
    }

    const v = s[k[k.length - 1]]
    return v !== undefined ? v : d
  }

  attachment (k, d) {
    const src = this.get(k, d)

    if (localAttachment.test(src)) {
      const [, attachmentID] = localAttachment.exec(src)

      return this.api.baseURL +
        this.api.attachmentOriginalEndpoint({
          attachmentID,
          kind: 'settings',
          name: k,
        })
    }

    if (src) {
      return this.api.baseURL
        .replace(/\/system$/, '')
        .replace(/\/api$/, '') + src
    }

    return d
  }
}

export default {
  install (app) {
    const settings = reactive(new Settings())
    app.config.globalProperties.$Settings = settings
    app.config.globalProperties.$s = settings.get.bind(settings)
  },
}
