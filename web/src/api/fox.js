
const api = require('./net').default

const Fox = {
  currency: async function () {
    return await api.get('https://api.gbi.news/currency', {})
  },
  b1Ticker: async function (name) {
    return await api.get(`/bigone/api/v3/asset_pairs/${name}/ticker`, {})
  }
};

export default Fox;
