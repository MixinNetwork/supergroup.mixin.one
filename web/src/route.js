import Vue from 'vue'
import VueRouter from 'vue-router'
import Home from './pages/Home'
import Auth from './pages/Auth'
import Blocking from './pages/Blocking'
import Pay from './pages/Pay'
import Broadcaster from './pages/Broadcaster'
import PreparePacket from './pages/PreparePacket'
import Packet from './pages/Packet'
import Members from './pages/Members'
import Messages from './pages/Messages'
import PageNotFound from './pages/PageNotFound'
import { ROUTER_MODE } from '@/constants.js'

Vue.use(VueRouter)

const routes = [
  { path: '/', component: Home },
  { path: '/pay', component: Pay },
  { path: '/blocking', component: Blocking },
  { path: '/broadcasters', component: Broadcaster },
  { path: '/packets/prepare', component: PreparePacket },
  { path: '/packets/:id', component: Packet },
  { path: '/members/', component: Members },
  { path: '/messages/', component: Messages },
  { path: '/auth', component: Auth },
  { path: "*", component: PageNotFound },
]

const router = new VueRouter({
  mode: ROUTER_MODE,
  routes // short for `routes: routes`
})

export default router
