import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'

import './console-splash'
import './themes'

if (!window.CortezaAPI) {
  window.CortezaAPI = '/api'
}
import('bootstrap/dist/css/bootstrap.min.css')
import('./config-check')

import routerConfig from './router'
import App from './App.vue'
import i18n from './i18n'

const app = createApp(App)

const pinia = createPinia()
app.use(pinia)

const router = createRouter({ history: createWebHistory(), routes: routerConfig })
app.use(router)
app.use(i18n)

import { plugins } from 'corteza-lib/vue/dist'
app.use(plugins.Auth({ app: 'unify', rootApp: true }))

import { CortezaAPI } from 'corteza-lib/vue/dist'
app.use(CortezaAPI('system'))
app.use(CortezaAPI('compose'))
app.use(CortezaAPI('automation'))

app.use(plugins.Settings, { api: app.config.globalProperties.$SystemAPI })

import { FontAwesomeIcon, FontAwesomeLayers } from '@fortawesome/vue-fontawesome'
import './components/faIcons'
app.component('font-awesome-icon', FontAwesomeIcon)
app.component('font-awesome-layers', FontAwesomeLayers)

import { filters } from 'corteza-lib/vue/dist'
for (const n in filters) { app.config.globalProperties[`$${n}`] = filters[n]
  app.config.globalProperties[n] = filters[n] }

const gp = app.config.globalProperties
const provides = ['$ComposeAPI', '$SystemAPI', '$Settings', '$auth']
provides.forEach(k => { if (gp[k]) app.provide(k, gp[k]) })
app.provide('can', () => true)
window.__auth = gp.$auth
window.__settings = gp.$Settings
window.__systemAPI = gp.$SystemAPI
window.__composeAPI = gp.$ComposeAPI
window.__automationAPI = gp.$AutomationAPI

app.mount('#app')
