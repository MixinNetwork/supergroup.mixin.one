
const api = require('./net').default

const Plugin = {
  shortcuts: async function () {
    return await api.get('/shortcuts', {})
  },
  redirect: async function (groupId, itemId) {
    return await api.get(`/shortcuts/${groupId}/${itemId}/redirect`)
  }
}

export default Plugin;
