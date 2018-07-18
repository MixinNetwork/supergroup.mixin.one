function Website(api) {
  this.api = api;
}

Website.prototype = {
  amount: function (callback) {
    this.api.request('GET', '/amount', undefined, function(resp) {
      return callback(resp);
    });
  }
};

export default Website;
