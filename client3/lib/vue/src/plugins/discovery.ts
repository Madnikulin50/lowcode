import { apiClients } from '@cortezaproject/corteza-js'
import { type App, type Plugin } from 'vue'

interface Options {
  baseURL?: string;
  accessTokenFn?: () => string | undefined;
}

export default function (opt: Options = {}): Plugin {
  return {
    install(app: App): void {
      if (!opt.baseURL) {
        const { CortezaDiscoveryAPI = undefined } = window as Record<string, string | undefined>
        if (CortezaDiscoveryAPI) {
          opt.baseURL = `${CortezaDiscoveryAPI}/`
        } else {
          opt.baseURL = 'http://localhost:3200/'
        }
      }

      if (!opt.accessTokenFn) {
        const auth = app.config.globalProperties.$auth
        if (auth && auth.accessTokenFn) {
          opt.accessTokenFn = auth.accessTokenFn
        }
      }

      if (opt.baseURL) {
        app.config.globalProperties.$DiscoveryAPI = new apiClients.Discovery(opt)
      } else {
        console.warn('window.CortezaDiscoveryAPI not set, $DiscoveryAPI not initialized')
      }
    },
  }
}
