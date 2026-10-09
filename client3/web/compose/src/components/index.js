import { FontAwesomeIcon, FontAwesomeLayers } from '@fortawesome/vue-fontawesome'
import PortalVue from 'portal-vue'
import './faIcons'
import { components } from 'corteza-lib/vue/dist'
import { CHelpTrigger, CHelpPanel, CHelpEditor } from 'corteza-lib/vue/src/components/help'
import CInputSelect from 'corteza-lib/vue/src/components/input/CInputSelect.vue'

import ECharts from 'vue-echarts'
import { use } from 'echarts/core'
import {
  CanvasRenderer,
} from 'echarts/renderers'
import {
  LineChart,
  BarChart,
  PieChart,
  GaugeChart,
  RadarChart,
  FunnelChart,
  ScatterChart,
  SankeyChart,
  HeatmapChart,
  BoxplotChart,
  CandlestickChart,
  GraphChart,
  MapChart,
  SunburstChart,
  ParallelChart,
  CustomChart,
} from 'echarts/charts'
import {
  TitleComponent,
  GridComponent,
  LegendComponent,
  TooltipComponent,
  VisualMapComponent,
  ToolboxComponent,
  DataZoomComponent,
  CalendarComponent,
  ParallelComponent,
} from 'echarts/components'

use([
  BarChart,
  LineChart,
  PieChart,
  GaugeChart,
  RadarChart,
  FunnelChart,
  ScatterChart,
  SankeyChart,
  HeatmapChart,
  BoxplotChart,
  CandlestickChart,
  GraphChart,
  MapChart,
  SunburstChart,
  ParallelChart,
  CustomChart,
  CanvasRenderer,
  TitleComponent,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  VisualMapComponent,
  ToolboxComponent,
  DataZoomComponent,
  CalendarComponent,
  ParallelComponent,
])

export default {
  install(app) {
    app.component('e-charts', ECharts)

    app.use(PortalVue)
    app.component('font-awesome-icon', FontAwesomeIcon)
    app.component('font-awesome-layers', FontAwesomeLayers)

    app.component('c-permissions-button', components.CPermissionsButton)
    app.component('c-input-confirm', components.CInputConfirm)
    app.component('c-input-processing', components.CInputProcessing)
    app.component('c-resource-list', components.CResourceList)
    app.component('c-input-checkbox', components.CInputCheckbox)
    app.component('c-button-submit', components.CButtonSubmit)
    app.component('c-hint', components.CHint)
    app.component('c-help-trigger', CHelpTrigger)
    app.component('c-help-panel', CHelpPanel)
    app.component('c-help-editor', CHelpEditor)
    app.component('c-input-select', CInputSelect)
    app.component('c-form-table-wrapper', components.CFormTableWrapper)
    app.component('c-webcam', components.CWebcam)
  },
}
