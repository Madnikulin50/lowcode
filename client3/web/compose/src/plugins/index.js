import BootstrapVue from 'bootstrap-vue'
import Router from 'vue-router'
import VueNativeSock from 'vue-native-websocket'

import { plugins, websocket } from 'corteza-lib/vue/dist'

import pairs from './eventbus-pairs'

const notProduction = (process.env.NODE_ENV !== 'production')
const verboseUIHooks = window.location.search.includes('verboseUIHooks')
const verboseEventbus = window.location.search.includes('verboseEventbus')

export default {
  install(app) {
    app.use(BootstrapVue, {
      BToast: {
        autoHideDelay: 7000,
        toaster: 'b-toaster-bottom-right',
      },
      BModal: {
        noEnforceFocus: true,
      },
    })
    app.use(Router)

    app.use(plugins.Auth(), { app: 'compose' })

    app.use(plugins.CortezaAPI('compose'))
    app.use(plugins.CortezaAPI('system'))
    app.use(plugins.CortezaAPI('federation'))
    app.use(plugins.CortezaAPI('automation'))
    app.use(plugins.DiscoveryAPI())

    app.use(plugins.EventBus(), {
      strict: notProduction,
      verbose: verboseEventbus,
      pairs,
    })

    app.use(plugins.UIHooks(), {
      app: 'compose',
      verbose: verboseUIHooks,
    })

    app.use(plugins.Settings, { api: app.config.globalProperties.$SystemAPI })
    app.use(plugins.Reminder, { api: app.config.globalProperties.$SystemAPI })

    app.use(VueNativeSock, websocket.endpoint(), websocket.config)
  },
}
