import { createRouter, createWebHistory } from 'vue-router'
import Home from './pages/HomePage'

const routes = [
  { path: '/', component: Home },
]

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes, // `routes: routes` 的缩写
})

export default router
