const api = require('./net').default

let Message = {
  index: async function () {
    return await api.get('/messages', {})
  },

  recall: async function (messageId) {
    return await api.post('/messages/' + messageId +'/recall', {}, {})
  }
}

export default Message
