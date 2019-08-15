const api = require('./net').default

const Account = {
  me: async function () {
    return await api.get('/me', {})
  },

  subscribe: async function () {
    return await api.post('/subscribe', {}, {})
  },

  unsubscribe: async function () {
    return await api.post('/unsubscribe', {}, {})
  },

  subscribers: async function (t=0, q='') {
    return await api.get('/subscribers?offset=' + t + '&q=' + q, {})
  },

  remove: async function (id) {
    return await api.post('/users/'+id+'/remove', {}, {})
  },

  block: async function (id) {
    return await api.post('/users/'+id+'/block', {}, {})
  },

  authenticate: async function (authorizationCode) {
    var params = {
      "code": authorizationCode
    };
    let resp = await api.post('/auth', params, {})
    if (resp.data) {
      window.localStorage.setItem('token', resp.data.authentication_token);
      window.localStorage.setItem('user_id', resp.data.user_id);
      window.localStorage.setItem('role', resp.data.role);
    }
    return resp
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
