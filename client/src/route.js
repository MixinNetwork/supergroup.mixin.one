import { createRouter, createWebHistory } from 'vue-router'
import Home from './pages/HomePage'
import PageNotFound from './pages/PageNotFound'

const routes = [
  { path: '/', component: Home },
  { path: "*", component: PageNotFound },
]

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes, // `routes: routes` 的缩写
})

export default router
