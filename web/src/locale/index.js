import Polyglot from 'node-polyglot';
const en = require('./en-US.json');
const zh = require('./zh-CN.json');

function Locale(lang) {
  var localeMap = {
    "en-US": require('./en-US.json'),
    "zh-CN": require('./zh-CN.json'),
  }
  var locale = LOCALE;
  this.polyglot = new Polyglot({locale: locale});
  this.polyglot.extend(localeMap[locale]);
}

Locale.prototype = {
  t: function(key, options) {
    return this.polyglot.t(key, options);
  }
};

export default Locale;
