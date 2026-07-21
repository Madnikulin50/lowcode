import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import PortalVue from 'portal-vue'

import 'bootstrap/dist/css/bootstrap.min.css'
import './config-check'
import './console-splash'
import './themes'

import routerConfig from './router'
import App from './App.vue'
import i18n from './i18n'

const app = createApp(App)

const pinia = createPinia()
app.use(pinia)

const router = createRouter({
  history: createWebHistory(),
  routes: routerConfig,
})
app.use(router)

app.use(i18n)

app.use(PortalVue)

// Auth
import { plugins } from 'corteza-lib/vue/dist'
app.use(plugins.Auth({ app: 'workflow' }))

// APIs
import { CortezaAPI } from 'corteza-lib/vue/dist'
app.use(CortezaAPI('system'))
app.use(CortezaAPI('compose'))
app.use(CortezaAPI('automation'))

// Settings
app.use(plugins.Settings, { api: app.config.globalProperties.$SystemAPI })

// FontAwesome
import { FontAwesomeIcon, FontAwesomeLayers } from '@fortawesome/vue-fontawesome'
import './components/faIcons'
app.component('font-awesome-icon', FontAwesomeIcon)
app.component('font-awesome-layers', FontAwesomeLayers)

// Lib components
import { components } from 'corteza-lib/vue/dist'
app.component('c-permissions-button', components.CPermissionsButton)
app.component('c-input-confirm', components.CInputConfirm)
app.component('c-button-submit', components.CButtonSubmit)
app.component('c-input-select', components.CInputSelect)

// Filters
import { filters } from 'corteza-lib/vue/dist'
for (const n in filters) {
  app.config.globalProperties[`$${n}`] = filters[n]
  app.config.globalProperties[n] = filters[n]
}

// Websocket
import { websocket } from 'corteza-lib/vue/dist'

const gp = app.config.globalProperties
const provides = [
  ['$ComposeAPI', 'composeAPI'],
  ['$SystemAPI', 'systemAPI'],
  ['$AutomationAPI', 'automationAPI'],
  ['$Settings'],
  ['$auth'],
]
provides.forEach(([dollar, plain]) => {
  if (gp[dollar]) app.provide(dollar, gp[dollar])
  if (plain && gp[dollar]) app.provide(plain, gp[dollar])
})
app.provide('can', () => true)
window.__systemAPI = gp.$SystemAPI
window.__composeAPI = gp.$ComposeAPI
window.__automationAPI = gp.$AutomationAPI
window.__auth = gp.$auth
window.__settings = gp.$Settings

app.mount('#app')
