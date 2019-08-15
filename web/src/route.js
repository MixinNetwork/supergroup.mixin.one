import Vue from 'vue'
import VueRouter from 'vue-router'
import Home from './pages/Home'
import TestAuth from './pages/TestAuth'
import Pay from './pages/Pay'
import PreparePacket from './pages/PreparePacket'
import Packet from './pages/Packet'
import Members from './pages/Members'
import Messages from './pages/Messages'
import { ROUTER_MODE } from '@/constants.js'

Vue.use(VueRouter)

const routes = [
  { path: '/', component: Home },
  { path: '/pay', component: Pay },
  { path: '/packets/prepare', component: PreparePacket },
  { path: '/packets/:id', component: Packet },
  { path: '/members/', component: Members },
  { path: '/messages/', component: Messages },
  { path: '/auth', component: TestAuth },
]

const router = new VueRouter({
  mode: ROUTER_MODE,
  routes // short for `routes: routes`
})

export default router
