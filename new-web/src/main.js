import Vue from 'vue'
import App from './App'
import router from './route'
import i18n from './i18n'
import global from './global'

import '@/plugins/vant'
import '@/plugins/vue-qr'
import '@/plugins/infinite-loading'

Vue.config.productionTip = false
Vue.prototype.GLOBAL = global
// console.log(global)
new Vue({
  components: {App},
  router,
  i18n,
  template: '<App/>'
}).$mount('#app')
