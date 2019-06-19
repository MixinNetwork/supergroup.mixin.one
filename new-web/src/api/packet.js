function Packet(api) {
  this.api = api;
}

Packet.prototype = {
  prepare: function (callback) {
    this.api.request('GET', '/packets/prepare', undefined, function(resp) {
      if (!!resp.error && resp.error.code === 403) {
        resp.error.code = 401;
        return
      }
      return callback(resp);
    });
  },

  create: function (callback, params) {
    this.api.request('POST', '/packets', params, function(resp) {
      return callback(resp);
    });
  },

  show: function (callback, packetId) {
    this.api.request('GET', '/packets/' + packetId, undefined, function(resp) {
      return callback(resp);
    });
  },

  claim: function (callback, packetId) {
    this.api.request('POST', '/packets/' + packetId + '/claim', undefined, function(resp) {
      return callback(resp);
    });
  }
};

export default Packet;
