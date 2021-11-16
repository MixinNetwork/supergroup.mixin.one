import Vue from 'vue'
import App from './App'
import router from './route'
import i18n from './i18n'
import global from './global'

//import '@/plugins/vant'
//import '@/plugins/vue-qr'
//import '@/plugins/infinite-loading'

//Vue.prototype.GLOBAL = global
/*
Vue.config.productionTip = false
new Vue({
  components: {App},
  router,
  i18n,
  template: '<App/>'
}).$mount('#app')
*/

const app = Vue.createApp({})
app.mount('#app')
