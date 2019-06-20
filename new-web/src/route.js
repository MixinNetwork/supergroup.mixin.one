import Vue from 'vue'
import VueRouter from 'vue-router'
import Home from './pages/Home'
import TestAuth from './pages/TestAuth'
import Pay from './pages/Pay'
import PreparePacket from './pages/PreparePacket'
import Packet from './pages/Packet'

Vue.use(VueRouter)

const routes = [
  { path: '/', component: Home },
  { path: '/pay', component: Pay },
  { path: '/packets/prepare', component: PreparePacket },
  { path: '/packets/:id', component: Packet },
  { path: '/auth', component: TestAuth }
]

const router = new VueRouter({
  routes // short for `routes: routes`
})

export default router