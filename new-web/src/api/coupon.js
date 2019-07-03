const api = require('./net').default

const Coupon = {
  index: async function () {
    return await api.get('/coupon', {})
  },

  create: async function (params) {
    return await api.post('/coupons', params, {})
  },

  occupy: async function (code) {
    return await api.post('/coupons/' + code, {}, {})
  }
}
export default Coupon;
