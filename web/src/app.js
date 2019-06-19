import 'simple-line-icons/scss/simple-line-icons.scss';
import './layout.scss';
import $ from 'jquery';
import Navigo from 'navigo';
import API from './api';
import Locale from './locale';
import Auth from './auth';
import User from './user';
import Packet from './packet';
import Message from './message';

const PartialLoading = require('./loading.html');
const Error404 = require('./404.html');
const router = new Navigo(WEB_ROOT, true);
const api = new API(router, API_ROOT);

window.i18n = new Locale(navigator.language);

router.replace = function(url) {
  this.resolve(url);
  this.pause(true);
  this.navigate(url);
  this.pause(false);
};

router.hooks({
  before: function(done, params) {
    $('body').attr('class', 'loading layout');
    $('#layout-container').html(PartialLoading());
    $('title').html(APP_NAME);
    done(true);
  },
  after: function(params) {
    router.updatePageLinks();
  }
});

router.on({
  '/auth': function () {
    new Auth(router, api).render();
  },
  '/': function () {
    new User(router, api).me();
  },
  '/subscribers': function () {
    new User(router, api).subscribers();
  },
  '/prepare': function () {
    new Packet(router, api).index();
  },
  '/packets/:id': function (params) {
    new Packet(router, api).show(params['id']);
  },
  '/messages': function () {
    new Message(router, api).index();
  }
}).notFound(function () {
  $('#layout-container').html(Error404());
  $('body').attr('class', 'error layout');
  router.updatePageLinks();
}).resolve();
