import axios from 'axios'
import { BASE_URL } from '@/constants'
import { Toast } from 'vant';

const instance = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
  responseType: 'json',
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json;charset=utf-8'
  }
})

instance.interceptors.response.use(response => {
  const { data: { code, data } } = response
  // handle custom error
  if (code === 0) {
    return Promise.resolve(data);
  } else {
    return Promise.reject(data)
  }
}, error => {
    if (error.response.status) {
      switch (error.response.status) {
        // handle http error
        case 404:
          Toast.fail('Not Exist')
          break
        default: 
          Toast.fail('Request error')
      }
    }
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