import { createRouter, createWebHistory } from 'vue-router'
import Home from './pages/HomePage'
//import Auth from './pages/Auth'
//import Blocking from './pages/Blocking'
//import Pay from './pages/Pay'
//import Broadcaster from './pages/Broadcaster'
//import PreparePacket from './pages/PreparePacket'
//import Packet from './pages/Packet'
import Members from './pages/MembersPage'
import Messages from './pages/MessagesPage'
//import PageNotFound from './pages/PageNotFound'

const routes = [
  { path: '/', component: Home },
  { path: '/members/', component: Members },
  { path: '/messages/', component: Messages },
]

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes, // `routes: routes` 的缩写
})

export default router
