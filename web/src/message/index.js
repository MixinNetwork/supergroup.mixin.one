import './index.scss';
import $ from 'jquery';

function Message(router, api) {
  this.router = router;
  this.api = api;
  this.templateIndex = require('./index.html');
}

Message.prototype = {
  index: function () {
    const self = this;
    self.api.message.index(function (resp) {
      if (resp.error) {
        return;
      }
      $('body').attr('class', 'message layout');
      var data = resp.data.map(msg => {
        msg.short_data = msg.data.slice(0, 100)
        return msg;
      });
      $('#layout-container').html(self.templateIndex({messages: data}));
      $(".messages").on('click', 'li', function (e) {
        e.preventDefault();
        if ($(this).data('category') == 'MESSAGE_RECALL') {
          return
        }
        $('.fullname').html($(this).data('name') + ' / ' + $(this).data('category'));
        $('.modal-name').html($(this).data('body'));
        $('.action.kick').data('id', $(this).data('id'));
        $('.modal-container').show();
      });
      $('.action.kick').on('click', function () {
        var that = this;
        $(".modal-container").hide();
        self.api.message.recall(function (resp) {
          if (resp.error) {
            return;
          }

          $("."+$(that).data('id')).hide();
        }, $(that).data('id'));
      });
      $('.icon-close').on('click', function () {
        $(".modal-container").hide();
      });
    });
  }
};

export default Message;
