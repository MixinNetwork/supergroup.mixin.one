function iosCopyToClipboard(el) {
  var oldContentEditable = el.contentEditable,
      oldReadOnly = el.readOnly,
      range = document.createRange();

  el.contentEditable = true;
  el.readOnly = false;
  range.selectNodeContents(el);

  var s = window.getSelection();
  s.removeAllRanges();
  s.addRange(range);

  el.setSelectionRange(0, 999999); // A big number, to cover anything that could be inside the element.

  el.contentEditable = oldContentEditable;
  el.readOnly = oldReadOnly;

  document.execCommand('copy')
}

export default {
  copyEl: function (el) {
    let text = el.innerText
    if (window) {
      if (navigator.userAgent.match(/ipad|ipod|iphone/i)) {
        iosCopyToClipboard(el)
      } else if (window.clipboardData && window.clipboardData.setData) {
          // IE specific code path to prevent textarea being shown while dialog is visible.
          try {
            window.clipboardData.setData("Text", text)
            return Promise.resolve()
          } catch (e) {
            return Promise.reject(e)
          }
      } else if (document.queryCommandSupported && document.queryCommandSupported("copy")) {
        var textarea = document.createElement("textarea")
        textarea.textContent = text
        textarea.style.position = "fixed";  // Prevent scrolling to bottom of page in MS Edge.
        document.body.appendChild(textarea)
        textarea.select()
        try {
            document.execCommand("copy");  // Security exception may be thrown by some browsers.
            return Promise.resolve()
        } catch (ex) {
            console.warn("Copy to clipboard failed.", ex)
            return Promise.reject(ex)
        } finally {
            document.body.removeChild(textarea)
        }
      }
    }
  },
  reloadPage: function () {
    let url = window.location.href
    var key = 't='
    var reg = new RegExp(key + '\\d+')
    var timestamp =+ new Date()
    let newUrl = url
    if(url.indexOf(key)>-1){
      // found timestamp
      newUrl = url.replace(reg, key + timestamp)
    } else {
      // no timestamp
      if (url.indexOf('?')>-1){
        var urlArr=url.split('?');
        if(urlArr[1]){
          newUrl = urlArr[0]+'?'+key+timestamp+'&'+urlArr[1]
        } else {
          newUrl = urlArr[0]+'?'+key+timestamp;
        }
      } else {
        if (url.indexOf('#')>-1){
          newUrl = url.split('#')[0]+'?'+key+timestamp+location.hash;
        } else {
          newUrl = url+'?'+key+timestamp;
        }
      }
    }
    window.location.href = newUrl
  }
}
