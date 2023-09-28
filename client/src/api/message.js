import api from './net'

const Message = {
  index: async function () {
    return await api.get('/messages', {})
  },

  recall: async function (messageId) {
    return await api.post('/messages/' + messageId +'/recall', {}, {})
  },

  add_featured_message: async function (messageId) {
    return await api.post('/featured_messages/' + messageId, {}, {})
  },

  featured_messages: async function () {
    return await api.get('/featured_messages', {})
  },
}

export default Message
