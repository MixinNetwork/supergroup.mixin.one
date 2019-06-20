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
  // { path: '/transactions/:id', component: Transaction },
  // { path: '/deposit', component: Deposit },
  // { path: '/withdraw', component: Withdraw },
  // { path: '/deposit/:assetId', component: Deposit },
  // { path: '/withdraw/:assetId', component: Withdraw },
  // { path: '/addresses/:assetId', component: AddressList },
  // { path: '/addresses/:assetId/create', component: EditAddress },
  // { path: '/addresses/:assetId/:id', component: EditAddress },
  { path: '/auth', component: TestAuth }
]

const router = new VueRouter({
  routes // short for `routes: routes`
})

export default router