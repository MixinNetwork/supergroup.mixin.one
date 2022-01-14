import api from './net'

const Website = {
  amount: async function () {
    return await api.get('/amount', {})
  },
  config: async function () {
    let resp = await api.get('/config', {})
    if (resp.data) {
      window.localStorage.setItem('cfg_client_id', resp.data.mixin_client_id);
      window.localStorage.setItem('cfg_host', resp.data.host);
    }
    return resp
  }
};

export default Website;
