import { createRouter, createWebHistory } from 'vue-router'
import Home from './pages/HomePage'
import Auth from './pages/AuthPage'
import Blocking from './pages/BlockingPage'
import Pay from './pages/PayPage'
import Proof from './pages/ProofPage'
import Broadcaster from './pages/BroadcasterPage'
import PreparePacket from './pages/PreparePacket'
import Packet from './pages/PacketPage'
import Members from './pages/MembersPage'
import Messages from './pages/MessagesPage'
import FeaturedMessages from './pages/FeaturedMessagesPage'
import PageNotFound from './pages/PageNotFound'

const routes = [
  { path: '/', component: Home },
  { path: '/auth', component: Auth },
  { path: '/broadcasters', component: Broadcaster },
  { path: '/blocking', component: Blocking },
  { path: '/pay', component: Pay },
  { path: '/proof', component: Proof },
  { path: '/packets/prepare', component: PreparePacket },
  { path: '/packets/:id', component: Packet },
  { path: '/members/', component: Members },
  { path: '/messages/', component: Messages },
  { path: '/featured_messages/', component: FeaturedMessages },
  { path: '/:pathMatch(.*)*', component: PageNotFound },
]

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes, // `routes: routes` 的缩写
})

export default router
