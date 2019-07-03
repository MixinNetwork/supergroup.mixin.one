import axios from 'axios'
import { BASE_URL } from '@/constants'

let headers = {
  'Content-Type': 'application/json',
  'Accept-Language': 'en-cn'
}

let instance = axios.create({
  timeout: 20000,
  withCredentials: true,
  headers
})

let HANDLERS = {
  401: [],
  500: []
}

const API = {
  /**
   * 配置参数，设置 headers，等信息
   * @param options
   *
   */
  config (options) {
    let { headers } = options
    Object.assign(instance.defaults.headers, headers)
  },
  /**
   * 用来绑定 权限校验失败、服务器异常、等事件
   * 当调用 axios 方法 或者权限校验出现上述异常时，触发对应的 handler
   * @param event
   * @param handler
   */
  on (event, handler) {
    let handlers = HANDLERS[event]
    if (!handlers) HANDLERS[event] = handlers = []
    handlers.push(handler)
  },

  trigger (event, payload) {
    let handlers = HANDLERS[event]
    if (handlers) handlers.forEach(hand => hand(payload))
  },

  request (options) {
    options.url = BASE_URL + options.url;
    let headers = options.headers || {}
    let token = options.token || window.localStorage.getItem('token')
    options.headers = Object.assign(headers, {'Authorization': 'Bearer ' + token })

    /* eslint prefer-promise-reject-errors: 0 */
    return instance.request(options).then(res => {
      // 如果设置不需要转换，则直接返回 res
      if (options.$parsed === false) return res
      if (!res.data) {
        return Promise.reject({
          code: 'response_error',
          message: 'response error',
          response: res
        })
      }
      let data = res.data
      if (!data) {
        return Promise.reject({
          code: -1,
          message: 'invalid data',
          response: res
        })
      }
      if (data.error) {
        let {code, description, status} = data.error
        let payload = {
          code,
          description: description,
          status: status,
          response: res
        }
        this.trigger(code, payload)
        return Promise.reject(payload)
      }
      return data
    }, ({response, message}) => {
      // @TODO
      console.log(response, message)
      if (!response) {
        let err = {code: 'network_error', message: 'network error'}
        return Promise.reject(err)
      }
    })
  },
  get (url, config = {}) {
    config.url = url
    config.method = 'get'
    return this.request(config)
  },
  post (url, data, config = {}) {
    config.url = url
    config.method = 'post'
    config.data = data
    return this.request(config)
  },
  put (url, data, config = {}) {
    config.url = url
    config.method = 'put'
    config.data = data
    return this.request(config)
  },
  delete (url, config = {}) {
    config.url = url
    config.method = 'delete'
    return this.request(config)
  },
};

export default API;
