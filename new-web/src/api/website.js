
const api = require('./net').default

const Website = {
  amount: async function () {
    return await api.get('/amount', {})
  }
};

export default Website;
