const api = require('./net').default

let Packet = {
  async prepare () {
    return await api.get('/packets/prepare', {});
  },

  async create (params) {
    return await api.post('/packets', params, {});
  },

  async show (packetId) {
    return await api.get('/packets/' + packetId, {})
  },

  async claim (packetId) {
    return await api.post('/packets/' + packetId + '/claim', {}, {});
  }
};

export default Packet;
