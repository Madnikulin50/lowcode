import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
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
const router = createRouter({ history: createWebHistory(), routes: routerConfig })
app.use(router)
app.use(i18n)

import { plugins } from 'corteza-lib/vue/dist'
app.use(plugins.Auth({ app: 'discovery' }))
import { CortezaAPI } from 'corteza-lib/vue/dist'
app.use(CortezaAPI('system'))
app.use(CortezaAPI('compose'))
app.use(plugins.Settings, { api: app.config.globalProperties.$SystemAPI })

import searcher from './plugins/searcher'
app.use(searcher())

import { FontAwesomeIcon, FontAwesomeLayers } from '@fortawesome/vue-fontawesome'
import './components/faIcons'
app.component('font-awesome-icon', FontAwesomeIcon)
app.component('font-awesome-layers', FontAwesomeLayers)

import { components } from 'corteza-lib/vue/dist'
app.component('c-permissions-button', components.CPermissionsButton)
app.component('c-input-confirm', components.CInputConfirm)
app.component('c-input-processing', components.CInputProcessing)
app.component('c-input-search', components.CInputSearch)

import VueSimpleMarkdown from 'vue-simple-markdown'
import 'vue-simple-markdown/dist/vue-simple-markdown.css'
app.use(VueSimpleMarkdown)

import { filters } from 'corteza-lib/vue/dist'
for (const n in filters) { app.config.globalProperties[`$${n}`] = filters[n]
  app.config.globalProperties[n] = filters[n] }

const gp = app.config.globalProperties
const provides = ['$ComposeAPI', '$SystemAPI', '$Settings', '$auth']
provides.forEach(k => { if (gp[k]) app.provide(k, gp[k]) })
app.provide('can', () => true)
window.__systemAPI = gp.$SystemAPI
window.__composeAPI = gp.$ComposeAPI
window.__auth = gp.$auth

app.mount('#app')