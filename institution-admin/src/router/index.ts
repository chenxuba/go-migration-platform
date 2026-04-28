import { createRouter, createWebHistory } from 'vue-router'
import staticRoutes from './static-routes'

const historyBase = import.meta.env.PROD ? '/institution/' : (import.meta.env.VITE_APP_BASE || '/')

const router = createRouter({
  routes: [
    ...staticRoutes,
  ],
  history: createWebHistory(historyBase),
})

export default router
