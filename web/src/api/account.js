function Account(api) {
  this.api = api;
}

Account.prototype = {
  me: function (callback) {
    this.api.request('GET', '/me', undefined, function(resp) {
      return callback(resp);
    });
  },

  subscribe: function (callback) {
    this.api.request('POST', '/subscribe', undefined, function(resp) {
      return callback(resp);
    });
  },

  unsubscribe: function (callback) {
    this.api.request('POST', '/unsubscribe', undefined, function(resp) {
      return callback(resp);
    });
  },

  subscribers: function (callback, t) {
    this.api.request('GET', '/subscribers?offset='+t, undefined, function(resp) {
      return callback(resp);
    });
  },

  remove: function (callback, id) {
    this.api.request('POST', '/users/'+id+'/remove', undefined, function(resp) {
      return callback(resp);
    });
  },

  block: function (callback, id) {
    this.api.request('POST', '/users/'+id+'/block', undefined, function(resp) {
      return callback(resp);
    });
  },

  authenticate: function (callback, authorizationCode) {
    var params = {
      "code": authorizationCode
    };
    this.api.request('POST', '/auth', params, function(resp) {
      if (resp.data) {
        window.localStorage.setItem('token', resp.data.authentication_token);
        window.localStorage.setItem('user_id', resp.data.user_id);
        window.localStorage.setItem('role', resp.data.role);
      }
      return callback(resp);
    });
  },

  userId: function () {
    return window.localStorage.getItem('user_id');
  },

  role: function () {
    return window.localStorage.getItem('role');
  },

  token: function () {
    return window.localStorage.getItem('token');
  },

  clear: function (callback) {
    window.localStorage.clear();
    if (typeof callback === 'function') {
      callback();
    }
  }
};

export default Account;
