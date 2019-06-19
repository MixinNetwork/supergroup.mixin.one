function Message(api) {
  this.api = api;
}

Message.prototype = {
  index: function (callback) {
    this.api.request('GET', '/messages', undefined, function(resp) {
      return callback(resp);
    });
  },

  recall: function (callback, messageId) {
    this.api.request('POST', '/messages/' +messageId+'/recall', undefined, function(resp) {
      return callback(resp);
    });
  }
}

export default Message;
