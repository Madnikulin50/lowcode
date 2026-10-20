import { createApp } from 'vue'
import './styles/global.css'
import App from './App.vue'
import IdCheckApp from './idcheck/IdCheckApp.vue'
import RemarksApp from './remarks/RemarksApp.vue'
import ObjectApp from './object/ObjectApp.vue'
import DocApp from './doc/DocApp.vue'

// ?view=idcheck — in-place findings viewer for the «Проверка ИД (АОСР)»
// namespace (embedded in the package/act cards); ?view=remarks — the workbench
// of the expertise remarks of an object; ?view=object — documents of an object
// (sets, completeness, which versions are compared); ?view=doc — one document with the assistant; otherwise the ПД/РД comparison viewer.
const view = new URLSearchParams(location.search).get('view')
createApp(view === 'idcheck' ? IdCheckApp : view === 'remarks' ? RemarksApp : view === 'object' ? ObjectApp : view === 'doc' ? DocApp : App).mount('#app')
