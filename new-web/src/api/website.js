
const api = require('./net').default

const Website = {
  amount: async function () {
    return await api.get('/amount', {})
  },
  config: async function () {
    return await api.get('/config', {})
  }
};

export default Website;
