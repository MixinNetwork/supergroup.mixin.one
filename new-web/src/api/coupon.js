const api = require('./net').default

const Coupon = {
  index: async function () {
    return await api.get('/coupons', {})
  },

  create: async function () {
    console.log('Coupon: create coupons')
    return await api.post('/coupons', {}, {})
  },

  occupy: async function (code) {
    return await api.post('/coupons/' + code, {}, {})
  }
}
export default Coupon;
