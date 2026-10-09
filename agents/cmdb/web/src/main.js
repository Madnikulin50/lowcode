import { createApp } from 'vue'
import { createRouter, createWebHashHistory } from 'vue-router'
import 'bootstrap/dist/css/bootstrap.min.css'
import 'bootstrap-icons/font/bootstrap-icons.css'
import App from './App.vue'
import Dashboard from './views/Dashboard.vue'
import DeviceList from './views/DeviceList.vue'

const routes = [
  { path: '/', name: 'dashboard', component: Dashboard },
  { path: '/devices', name: 'devices', component: DeviceList },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

createApp(App).use(router).mount('#app')
