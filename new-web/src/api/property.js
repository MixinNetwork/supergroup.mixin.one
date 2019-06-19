const api = require('./net').default

let Property = {
  async create (state) {
    return await api.post('/properties', {'value': state}, {})
  }
}

export default Property;
