import { PluginFunction } from 'vue'
import { eventbus } from '../../../../lib/js/dist'

export default function (): PluginFunction<Partial<eventbus.Options>> {
  return function (Vue, opts): void {
    Vue.prototype.$EventBus = new eventbus.EventBus(opts)
  }
}
