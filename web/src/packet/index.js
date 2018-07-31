import './index.scss';
import $ from 'jquery';
import FormUtils from '../utils/form.js';
import TimeUtils from '../utils/time.js';

function Packet(router, api) {
  this.router = router;
  this.api = api;
  this.partialLoading = require('../loading.html');
  this.partialAsset = require('./asset.html');
  this.partialParticipant = require('./participant.html');
  this.templateEmpty = require('./empty.html');
  this.templateIndex = require('./index.html');
  this.templateShow = require('./show.html');
  this.templateOpen = require('./open.html');
  this.templateState = require('./state.html');
}

Packet.prototype = {
  index: function () {
    const self = this;
    self.api.packet.prepare(function (resp) {
      if (resp.error) {
        return;
      }
      var data = resp.data;
      $('body').attr('class', 'packet layout');
      if (data.assets.length === 0) {
        $('#layout-container').html(self.templateEmpty());
        return;
      }
      data['asset'] = self.selectAsset(data.assets);
      $('#layout-container').html(self.templateIndex(data));
      for (var i in data.assets) {
        var asset = data.assets[i];
        $('.assets.list.container').append(self.partialAsset(asset));
      }
      $('.assets.list.container').on('click', '.asset.item', function (event) {
        event.preventDefault();
        self.changeAsset($(this).data('asset-id'));
      });
      $('.assets.list.action').click(function (event) {
        event.preventDefault();
        $('.assets.list.container').slideToggle();
        $('.form.container').slideToggle();
      });
      self.handleCreate();
      self.router.updatePageLinks();
    });
  },

  show: function (packetId) {
    const self = this;
    self.api.packet.show(function (resp) {
      if (resp.error) {
        return;
      }
      var data = resp.data;
      if (data.state === 'INITIAL') {
        return;
      }
      $('body').attr('class', 'packet layout');
      for (var i in data.participants) {
        var p = data.participants[i];
        if (p.user_id === self.api.account.userId()) {
          data.lottery = p;
          break;
        }
      }
      data['firstLetter'] = data.user.avatar_url === '' ? (data.user.full_name.trim()[0] || '^_^') : undefined;
      if (data.lottery || data.state === 'EXPIRED' || data.state === 'REFUNDED') {
        $('#layout-container').html(self.templateShow(data));
        for (var i in data.participants) {
          var participant = data.participants[i];
          participant['symbol'] = data.asset.symbol;
          participant['created_at'] = TimeUtils.format(participant.created_at);
          participant['firstLetter'] = participant.avatar_url === '' ? (participant.full_name.trim()[0] || '^_^') : undefined;
          $('.packet.history ul').append(self.partialParticipant(participant));
        }
      } else {
        $('body').addClass('open');
        $('#layout-container').html(self.templateOpen(data));
        $('.packet.open.button a').click(function (event) {
          event.preventDefault();
          self.handleClaim(packetId);
        });
      }
      self.router.updatePageLinks();
    }, packetId);
  },

  handleClaim: function (packetId) {
    const self = this;
    $('.submitting.overlay').show();
    self.api.packet.claim(function (resp) {
      if (resp.error) {
        $('.submitting.overlay').hide();
        return;
      }
      var data = resp.data;
      for (var i in data.participants) {
        var p = data.participants[i];
        if (p.user_id === self.api.account.userId()) {
          data.lottery = p;
          break;
        }
      }
      $('body').removeClass('open');
      data['firstLetter'] = data.user.avatar_url === '' ? (data.user.full_name.trim()[0] || '^_^') : undefined;
      $('#layout-container').html(self.templateShow(data));
      for (var i in data.participants) {
        var participant = data.participants[i];
        participant['symbol'] = data.asset.symbol;
        participant['created_at'] = TimeUtils.format(participant.created_at);
        participant['firstLetter'] = participant.avatar_url === '' ? (participant.full_name.trim()[0] || '^_^') : undefined;
        $('.packet.history ul').append(self.partialParticipant(participant));
      }
    }, packetId);
  },

  handleCreate: function () {
    const self = this;
    $('form').submit(function (event) {
      event.preventDefault();
      var params = FormUtils.serialize($(this));
      params['total_count'] = parseInt(params.total_count);
      self.api.packet.create(function (resp) {
        if (resp.error) {
          $('.submitting.overlay').hide();
          return;
        }
        var pkt = resp.data;
        console.info('mixin://pay?recipient=' + CLIENT_ID + '&asset=' + pkt.asset.asset_id + '&amount=' + pkt.amount + '&trace=' + pkt.packet_id + '&memo=' + pkt.greeting);
        setTimeout(function() { self.waitForPayment(pkt.packet_id); }, 1500);
        window.location.replace('mixin://pay?recipient=' + CLIENT_ID + '&asset=' + pkt.asset.asset_id + '&amount=' + pkt.amount + '&trace=' + pkt.packet_id + '&memo=' + pkt.greeting);
      }, params);
    });
    $('input[type=submit]').click(function (event) {
      event.preventDefault();
      $('.submitting.overlay').show();
      $(this).parents('form').submit();
    });
  },

  waitForPayment: function (packetId) {
    const self = this;
    self.api.packet.show(function (resp) {
      if (resp.error) {
        setTimeout(function() { self.waitForPayment(packetId); }, 1500);
        return;
      }
      var data = resp.data;
      switch (data.state) {
        case 'INITIAL':
          setTimeout(function() { self.waitForPayment(packetId); }, 1500);
          break;
        case 'PAID':
        case 'EXPIRED':
        case 'REFUNDED':
          $('#layout-container').html(self.templateState(data));
          break;
      }
    }, packetId);
  },

  changeAsset: function (assetId) {
    const self = this;
    if (assetId === window.localStorage.getItem('asset_id')) {
      $('.assets.list.container').slideToggle();
      $('.form.container').slideToggle();
      return;
    }
    window.localStorage.setItem('asset_id', assetId);
    $('body').attr('class', 'loading layout');
    $('#layout-container').html(self.partialLoading());
    self.index();
  },

  selectAsset: function (assets) {
    var assetId = window.localStorage.getItem('asset_id');
    for (var i in assets) {
      if (assetId === assets[i].asset_id) {
        return assets[i];
      }
    }
    var asset = assets[0];
    window.localStorage.setItem('asset_id', asset.asset_id);
    return asset;
  }
};

export default Packet;
