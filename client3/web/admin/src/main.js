import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'

if (!window.CortezaAPI) {
  window.CortezaAPI = '/api'
}
import 'bootstrap/dist/css/bootstrap.min.css'
import('./config-check')
import('./console-splash')
import './themes'

import { useStore } from './store'
import routerConfig from './router'
import App from './App.vue'
import i18n from './i18n'

const app = createApp(App)

// Pinia store
const pinia = createPinia()
app.use(pinia)

// Router
const router = createRouter({
  history: createWebHistory(),
  routes: routerConfig,
})
app.use(router)

// i18n
app.use(i18n)

// Initialize auth
import { plugins } from 'corteza-lib/vue/dist'
app.use(plugins.Auth({ app: 'admin' }))

// APIs
import { CortezaAPI } from 'corteza-lib/vue/dist'
app.use(CortezaAPI('compose'))
app.use(CortezaAPI('system'))
app.use(CortezaAPI('federation'))
app.use(CortezaAPI('automation'))

// EventBus
app.use(plugins.EventBus(), { strict: true })

// UI Hooks
app.use(plugins.UIHooks(), { app: 'admin' })

// Settings
app.use(plugins.Settings, { api: app.config.globalProperties.$SystemAPI })

// FontAwesome
import { FontAwesomeIcon, FontAwesomeLayers } from '@fortawesome/vue-fontawesome'
import './components/faIcons'
app.component('font-awesome-icon', FontAwesomeIcon)
app.component('font-awesome-layers', FontAwesomeLayers)

// Lib components
import { components } from 'corteza-lib/vue/dist'
app.component('c-corredor-manual-buttons', components.CCorredorManualButtons)
app.component('c-permissions-button', components.CPermissionsButton)
app.component('c-input-confirm', components.CInputConfirm)
app.component('c-button-submit', components.CButtonSubmit)
app.component('c-input-checkbox', components.CInputCheckbox)
app.component('c-input-select', components.CInputSelect)
app.component('c-form-table-wrapper', components.CFormTableWrapper)

// ECharts
import ECharts from 'vue-echarts'
import { use } from 'echarts/core'
import { SVGRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import { TitleComponent, GridComponent, TooltipComponent } from 'echarts/components'
use([LineChart, SVGRenderer, TitleComponent, GridComponent, TooltipComponent])
app.component('e-charts', ECharts)

// Filters
import { filters } from 'corteza-lib/vue/dist'
for (const n in filters) {
  app.config.globalProperties[`$${n}`] = filters[n]
  app.config.globalProperties[n] = filters[n]
}

// App-specific component: CTheMainNav, CContentHeader, CSystemFields, CResourceListStatusFilter
import CTheMainNav from './components/CTheMainNav.vue'
import CContentHeader from './components/CContentHeader.vue'
import CSystemFields from './components/CSystemFields.vue'
import CResourceListStatusFilter from './components/CResourceListStatusFilter.vue'
app.component('c-the-main-nav', CTheMainNav)
app.component('c-content-header', CContentHeader)
app.component('c-system-fields', CSystemFields)
app.component('c-resource-list-status-filter', CResourceListStatusFilter)

// Websocket
import { websocket } from 'corteza-lib/vue/dist'
const wsEndpoint = websocket.endpoint()

const gp = app.config.globalProperties
const provides = ['$ComposeAPI', '$SystemAPI', '$Settings', '$auth']
provides.forEach(k => { if (gp[k]) app.provide(k, gp[k]) })
window.__systemAPI = gp.$SystemAPI
window.__SystemAPI = gp.$SystemAPI
window.__composeAPI = gp.$ComposeAPI
window.__ComposeAPI = gp.$ComposeAPI
window.__automationAPI = gp.$AutomationAPI
window.__AutomationAPI = gp.$AutomationAPI
window.__federationAPI = gp.$FederationAPI
window.__auth = gp.$auth
window.__settings = gp.$Settings

app.mount('#app')
