import axios from 'axios'
import { MIXIN_HOST } from '@/constants'
import { Toast } from 'vant'
import OAuth from '@/utils/oauth_pkce'

const instance = axios.create({
  baseURL: MIXIN_HOST,
  timeout: 10000,
  responseType: 'json',
  headers: {
    'Content-Type': 'application/json;charset=utf-8',
  }
})

instance.interceptors.request.use(
  config => {
    if (config.url === '/oauth/token'){
      return config
    }
    // if (process.env.NODE_ENV !== 'production') {
    //   config.headers.Authorization = 'Bearer eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJhaWQiOiJmNDAzYjk0OC02NzBkLTRkMjAtYTQwYy1kNDc4YzdiZTVkMWUiLCJleHAiOjE1OTA0NzQ5NzEsImlhdCI6MTU1ODkzODk3MSwiaXNzIjoiMTdhYzUyNWItNWUxMi00NGIwLThmNTEtNWJlYjhhYTFhMTI5In0.MxwEV5byAagS-nb28Kyq9EkdeqxcewL3n0ekRG2eSbZS1iyeUEXl5neqFOR1ylBwvJgh3J5H0chWZWYyGWdQxB2NLHP8aF8QoVudFJfpiv67Jf-5fwrmkcZHOLqRYyidDQdFrwOCpxnHKqPBzvFmVzknyqKce1EasL8SsrfmXnw'
    // }
    if (localStorage.getItem('token')) {
      const token = localStorage.getItem('token')
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  }
)

instance.interceptors.response.use(response => {
  const { data: { error } } = response
  if (!error) {
    const { data: { data } }= response
    return Promise.resolve(data);
  } else {
    const { code } = error
    switch (code) {
      // access token invalid
      case 401:
        OAuth.requestCode()
        break
      default:
        Toast('Request error')
    }
    return Promise.reject(error)
  }
}, () => {
  Toast('Request error')
})

async function request(options) {
  const res = await instance.request(options)
  return Promise.resolve(res)
}

export default {
  post(url, params = {}) {
    const options = {
      url,
      method: 'post',
      ...params
    }
    return request(options)
  },

  get(url, params = {}) {
    const options = {
      url,
      methods: 'get',
      ...params
    }
    return request(options)
  }
}