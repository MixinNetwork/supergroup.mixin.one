import api from './net'

const Broadcaster = {
  index: async function () {
    return await api.get('/broadcasters', {});
  },

  create: async function (identity) {
    return await api.post('/broadcasters', {'identity': parseInt(identity)}, {});
  },

  assets: async function () {
    return await api.get('/assets', {});
  }
}

export default Broadcaster;
