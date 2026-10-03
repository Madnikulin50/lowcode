import { createApp } from 'vue'
import './styles/global.css'
import App from './App.vue'
import IdCheckApp from './idcheck/IdCheckApp.vue'

// ?view=idcheck — in-place findings viewer for the «Проверка ИД (АОСР)»
// namespace (embedded in the package/act cards); otherwise the ПД/РД
// comparison viewer.
const view = new URLSearchParams(location.search).get('view')
createApp(view === 'idcheck' ? IdCheckApp : App).mount('#app')
