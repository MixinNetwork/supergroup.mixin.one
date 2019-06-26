import Vue from 'vue'
import VueRouter from 'vue-router'
import Home from './pages/Home'
import TestAuth from './pages/TestAuth'
import Pay from './pages/Pay'
import PayWxQr from './pages/PayWxQr'
import PreparePacket from './pages/PreparePacket'
import Packet from './pages/Packet'
import Members from './pages/Members'
import Coupons from './pages/Coupons'
import Messages from './pages/Messages'
import WxPay from './pages/WxPay'
import WxPayDone from './pages/WxPayDone'

Vue.use(VueRouter)

const routes = [
  { path: '/', component: Home },
  { path: '/pay', component: Pay },
  { path: '/pay/wxqr', component: PayWxQr },
  { path: '/packets/prepare', component: PreparePacket },
  { path: '/packets/:id', component: Packet },
  { path: '/members/', component: Members },
  { path: '/messages/', component: Messages },
  { path: '/coupons/', component: Coupons },
  { path: '/auth', component: TestAuth },
  // special route for wechat, no auth required.
  { path: '/wxpay', component: WxPay }, 
  { path: '/wxpay/done', component: WxPayDone }, 
]

const router = new VueRouter({
  routes // short for `routes: routes`
})

export default router