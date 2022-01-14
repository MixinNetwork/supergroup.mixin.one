import { createApp } from 'vue'
import App from './App.vue'
import Vant from 'vant';
import 'vant/lib/index.css';
import router from './route'
import i18n from './i18n'
import global from './global'

const app = createApp(App)
app.config.globalProperties.GLOBAL = global

app.use(Vant);

app.use(router)
app.use(i18n)
app.mount('#app')
