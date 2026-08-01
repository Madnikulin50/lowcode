import { createApp, h, defineComponent } from 'vue'
import { createPinia } from 'pinia'
import './themes'

import App from './App.vue'
import router from './router'
import i18n from './i18n'

if (!window.CortezaAPI) {
  window.CortezaAPI = '/api'
}

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(i18n)

import { Translation } from 'vue-i18n'

app.component('i18next', defineComponent({
  props: {
    path: { type: String, required: true },
    tag: { type: String, default: 'span' },
  },
  setup(props, { slots }) {
    return () => {
      const children = slots.default?.() || []
      const namedSlots = {}
      children.forEach((child, i) => {
        namedSlots[i] = () => child
      })
      return h(Translation, { keypath: props.path, tag: props.tag, scope: 'global' }, namedSlots)
    }
  },
}))

import { plugins } from 'corteza-lib/vue/dist'
app.use(plugins.Auth({ app: 'compose' }))
import { CortezaAPI } from 'corteza-lib/vue/dist'
app.use(CortezaAPI('system'))
app.use(CortezaAPI('compose'))
app.use(CortezaAPI('automation'))
app.use(plugins.Settings, { api: app.config.globalProperties.$SystemAPI })
app.use(plugins.EventBus(), { strict: true })
app.use(plugins.UIHooks(), { app: 'compose' })
app.use(plugins.Reminder, { api: app.config.globalProperties.$ComposeAPI })

import localComponents from './components'
app.use(localComponents)

import { createBootstrap } from 'bootstrap-vue-next'
import { BButton, BInputGroup } from 'bootstrap-vue-next/components'
import 'bootstrap/dist/css/bootstrap.css'
import 'bootstrap-vue-next/dist/bootstrap-vue-next.css'
app.use(createBootstrap())
app.component('BButton', BButton)
app.component('BInputGroup', BInputGroup)

import { components } from 'corteza-lib/vue/dist'
app.component('c-corredor-manual-buttons', components.CCorredorManualButtons)

import { filters } from 'corteza-lib/vue/dist'
for (const n in filters) {
  app.config.globalProperties[`$${n}`] = filters[n]
  app.config.globalProperties[n] = filters[n]
}

const gp = app.config.globalProperties
const provides = ['$ComposeAPI', '$SystemAPI', '$AutomationAPI', '$Settings', '$auth', '$EventBus', '$UIHooks', '$Reminder', '$DiscoveryAPI']
provides.forEach(k => {
  if (gp[k]) app.provide(k, gp[k])
})
window.__auth = gp.$auth
window.__settings = gp.$Settings
window.__systemAPI = gp.$SystemAPI
window.__composeAPI = gp.$ComposeAPI
window.__automationAPI = gp.$AutomationAPI
window.__eventBus = gp.$EventBus
window.__uiHooks = gp.$UIHooks
window.__Reminder = gp.$Reminder

app.mount('#app')
