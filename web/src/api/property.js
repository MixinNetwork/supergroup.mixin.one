import api from './net'

const Property = {
  async create (state) {
    return await api.post('/properties', {'value': state}, {})
  }
}

export default Property;
