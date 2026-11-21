import { createRouter, createWebHistory } from 'vue-router'
import routes from './views/routes'

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.onError((error) => {
  console.warn('Navigation error occurred:', error)
})

export default router
