import http from './http/mixin'

export default {
  getProfile() {
    return http.get('/me')
  }
}