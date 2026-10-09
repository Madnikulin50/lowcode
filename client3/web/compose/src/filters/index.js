import { filters } from 'corteza-lib/vue/dist'

export default {
  install(app) {
    for (const n in filters) {
      app.config.globalProperties[`$${n}`] = filters[n]
    }
  },
}
