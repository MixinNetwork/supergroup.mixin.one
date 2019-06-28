import './index.scss';
import $ from 'jquery';
import TimeUtils from '../utils/time.js';
import URLUtils from '../utils/url.js';
import MixinUtils from '../utils/mixin.js';

function User(router, api) {
  this.router = router;
  this.api = api;
  this.templateShow = require('./show.html');
  this.templatePayment = require('./payment.html');
  this.templateSubscribers = require('./subscribers.html');
}

User.prototype = {
  me: function (resp) {
    let self = this;
    self.api.account.me(function (resp) {
      if (resp.error) {
        return;
      }

      var data = resp.data;
      if (data.state == 'pending') {
        $('body').attr('class', 'user layout');
        $('#layout-container').html(self.templatePayment());
        $('.btn.pay').on('click', function() {
          $('.submitting.overlay').show();
          setTimeout(function() { self.waitForPayment(); }, 1500);
          window.location.replace('mixin://pay?recipient=' + CLIENT_ID + '&asset=c94ac88f-4671-3976-b60a-09064f1811e8&amount=0.05&trace=' + data.trace_id + '&memo=0.05%20XIN');
        });
        return;
      }

      self.api.website.statistics(function (resp) {
        if (resp.error) {
          return;
        }

        $('body').attr('class', 'user layout');
        $('#layout-container').html(self.templateShow({
          isMixin: !!MixinUtils.environment(),
          isAdmin: self.api.account.role() === 'admin'
        }));
        $('.members').html(window.i18n.t('user.participants.count') + resp.data.users_count);
        if (data.subscribed_at === "0001-01-01T00:00:00Z") {
          $('.subscribe').show();
        } else {
          $('.unsubscribe').show();
        }
        if (resp.data.prohibited) {
          $('.unprohibited').show();
        } else {
          $('.prohibited').show();
        }
        $('.subscribe').on('click', function (e) {
          e.preventDefault();
          if ($('.subscribe').hasClass('disabled')) {
            return;
          }
          $('.subscribe').addClass('disabled');
          self.api.account.subscribe(function (resp) {
            if (resp.error) {
              return;
            }
            $('.subscribe').removeClass('disabled');
            $('.subscribe').hide();
            $('.unsubscribe').show();
          });
        });
        $('.unsubscribe').on('click', function (e) {
          e.preventDefault();
          if ($('.unsubscribe').hasClass('disabled')) {
            return;
          }
          $('.unsubscribe').addClass('disabled');
          self.api.account.unsubscribe(function (resp) {
            if (resp.error) {
              return;
            }
            $('.unsubscribe').removeClass('disabled');
            $('.unsubscribe').hide();
            $('.subscribe').show();
          });
        });

        $('.prohibited').on('click', function (e) {
          e.preventDefault();
          if ($('.prohibited').hasClass('disabled')) {
            return;
          }
          $('.prohibited').addClass('disabled');
          self.api.property.create(function (resp) {
            if (resp.error) {
              return;
            }
            $('.prohibited').removeClass('disabled');
            $('.prohibited').hide();
            $('.unprohibited').show();
          }, true);
        });
        $('.unprohibited').on('click', function (e) {
          e.preventDefault();
          if ($('.unprohibited').hasClass('disabled')) {
            return;
          }
          $('.unprohibited').addClass('disabled');
          self.api.property.create(function (resp) {
            if (resp.error) {
              return;
            }
            $('.unprohibited').removeClass('disabled');
            $('.unprohibited').hide();
            $('.prohibited').show();
          }, false);
        });
      });
    });
  },

  subscribers: function () {
    let q = URLUtils.getUrlParameter('q');
    let self = this;
    self.api.account.subscribers(function (resp) {
      if (resp.error) {
        return;
      }

      var offset = '';
      if (resp.data.length > 0) {
        offset = resp.data[resp.data.length-1].subscribed_at;
      }
      var role = self.api.account.role();
      for (var i in resp.data) {
        var data  = resp.data[i];
        data['subscribed_at'] = TimeUtils.format(data.subscribed_at);
        if (role != 'admin') {
          data['identity_number'] = '';
        }
      }
      $('body').attr('class', 'user layout');
      $('#layout-container').html(self.templateSubscribers({admin: role == 'admin', subscribers: resp.data}));
      if (resp.data.length == 200) {
        $('.action.more').show();
        $('.action.more').on('click', function (e) {
          e.preventDefault();
          if ($('.action.more').hasClass('disabled')) {
            return;
          }

          $('.action.more').addClass('disabled');
          self.api.account.subscribers(function (resp) {
            if (resp.error) {
              return;
            }

            var html = '';
            if (resp.data.length > 1) {
              data = resp.data[resp.data.length-1];
              offset = data.subscribed_at;
            }
            for (var i in resp.data) {
              data  = resp.data[i];
              if (role != 'admin') {
                data['identity_number'] = '';
              }
              html += '<li class="'+data.user_id+'" data-id="'+data.user_id+'" data-display="'+data.full_name+' '+data.identity_number+'" data-avatar="'+data.avatar_url+'"><img src="'+data.avatar_url+'">'+data.full_name+' '+data.identity_number+'<span class="date">'+TimeUtils.format(data.subscribed_at)+'</span></li>';
            }
            if (resp.data.length < 200) {
              $('.action.more').hide();
            } else {
              $('.action.more').removeClass('disabled');
            }
            $('.list', '.subscribers').append(html);
          }, offset);
        });
      };

      $('.search.bar').on('click', '.icon', function (e) {
        e.preventDefault();

        window.location.replace('/subscribers?q='+$('.search.bar > .q').val());
      });
      if (self.api.account.role() === "admin") {
        $(".subscribers").on('click', 'li', function (e) {
          e.preventDefault();

          var that = this;
          $('.modal-avatar').attr('src', $(that).data('avatar'));
          $('.modal-name').html($(that).data('display'));
          $('.action.kick').data('id', $(that).data('id'));
          $('.action.block').data('id', $(that).data('id'));
          $('.modal-container').show();
        });
        $('.action.kick').on('click', function () {
          var that = this;
          $(".modal-container").hide();
          self.api.account.remove(function (resp) {
            if (resp.error) {
              return;
            }

            $("."+$(that).data('id')).hide();
          }, $(that).data('id'));
        });
        $('.action.block').on('click', function () {
          var that = this;
          $(".modal-container").hide();
          self.api.account.block(function (resp) {
            if (resp.error) {
              return;
            }

            $("."+$(that).data('id')).hide();
          }, $(that).data('id'));
        });
        $('.icon-close').on('click', function () {
          $(".modal-container").hide();
        });
      };
    }, "", q);
  },

  waitForPayment: function () {
    let self = this;
    self.api.account.me(function (resp) {
      if (resp.error) {
        setTimeout(function() { self.waitForPayment(); }, 1500);
        return;
      }

      if (resp.data.state === 'paid') {
        self.router.replace('/');
        return;
      }
      setTimeout(function() { self.waitForPayment(); }, 1500);
    })
  }
};

export default User;
