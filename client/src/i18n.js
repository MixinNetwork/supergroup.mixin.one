import { createI18n } from 'vue-i18n'

const i18n = createI18n({
  locale: window.navigator.language.indexOf('zh') !== -1 ? 'zh' : 'en',  // set locale
  fallbackLocale: 'en',
  messages: {
    en: require('../locales/en.json'),
    zh: require('../locales/zh.json')
  }
})

export default i18n
