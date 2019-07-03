import Vue from 'vue'
import VueI18n from 'vue-i18n'

Vue.use(VueI18n)

const i18n = new VueI18n({
  locale: window.navigator.language.indexOf('zh') !== -1 ? 'zh' : 'en',       // set locale
  fallbackLocale: 'en',
  messages: {
    en: require('../locales/en.json'),
    zh: require('../locales/zh.json')
  }
})

export default i18n