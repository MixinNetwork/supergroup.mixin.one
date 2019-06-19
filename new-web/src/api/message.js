const api = require('./net').default

let Message = {
  index: async function () {
    return await api.get('GET', '/messages', {})
  },

  recall: async function (messageId) {
    return await api.post('POST', '/messages/' + messageId +'/recall', {}, {})
  }
}

export default Message
