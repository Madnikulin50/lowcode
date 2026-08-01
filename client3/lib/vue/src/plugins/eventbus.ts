import { type App, type Plugin } from 'vue'
import { eventbus } from 'corteza-lib/js/dist'

export default function (opts?: Partial<eventbus.Options>): Plugin {
  return {
    install(app: App): void {
      app.config.globalProperties.$EventBus = new eventbus.EventBus(opts)
    },
  }
}
