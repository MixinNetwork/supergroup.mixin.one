function Property(api) {
  this.api = api;
}

Property.prototype = {
  create: function (callback, state) {
    this.api.request('POST', '/properties', {'value': state}, function(resp) {
      return callback(resp);
    })
  }
}

export default Property;
