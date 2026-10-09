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
app.use(plugins.Auth({ app: 'reporter' }))

import { CortezaAPI } from 'corteza-lib/vue/dist'
app.use(CortezaAPI('system'))
app.use(CortezaAPI('compose'))

app.use(plugins.Settings, { api: app.config.globalProperties.$SystemAPI })

import { FontAwesomeIcon, FontAwesomeLayers } from '@fortawesome/vue-fontawesome'
import './components/faIcons'
app.component('font-awesome-icon', FontAwesomeIcon)
app.component('font-awesome-layers', FontAwesomeLayers)

import { components } from 'corteza-lib/vue/dist'
app.component('c-permissions-button', components.CPermissionsButton)
app.component('c-input-confirm', components.CInputConfirm)
app.component('c-button-submit', components.CButtonSubmit)
app.component('c-input-select', components.CInputSelect)
app.component('c-form-table-wrapper', components.CFormTableWrapper)

import ECharts from 'vue-echarts'
import { use } from 'echarts/core'
import { SVGRenderer } from 'echarts/renderers'
import { LineChart, BarChart, PieChart, GaugeChart, RadarChart, FunnelChart } from 'echarts/charts'
import { TitleComponent, GridComponent, LegendComponent, TooltipComponent, VisualMapComponent, ToolboxComponent } from 'echarts/components'
use([BarChart, LineChart, PieChart, GaugeChart, RadarChart, FunnelChart, SVGRenderer, TitleComponent, GridComponent, TooltipComponent, LegendComponent, VisualMapComponent, ToolboxComponent])
app.component('e-charts', ECharts)

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
