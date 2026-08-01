import { apiClients } from 'corteza-lib/js/dist'
import { type App, type Plugin } from 'vue'

interface Options {
  baseURL?: string;
  accessTokenFn?: () => string | undefined;
}

export default function (service: string, opt: Options = {}): Plugin {
  return {
    install(app: App): void {
      if (!opt.baseURL) {
        if (!(window as Record<string, string | undefined>).CortezaAPI) {
          throw new Error('config.js missing or window.CortezaAPI not set')
        }

        opt.baseURL = `${(window as Record<string, string | undefined>).CortezaAPI}/${service}`
      }

      const capService = service.substring(0, 1).toUpperCase() + service.substring(1)

      if (!opt.accessTokenFn) {
        const auth = app.config.globalProperties.$auth
        if (auth && auth.accessTokenFn) {
          opt.accessTokenFn = auth.accessTokenFn
        }
      }

      app.config.globalProperties[`$${capService}API`] = new apiClients[capService](opt)
    },
  }
}
