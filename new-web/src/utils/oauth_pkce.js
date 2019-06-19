// 使用PKCE方式在客户端获取Access_token
// https://auth0.com/docs/api-auth/tutorials/authorization-code-grant-pkce
import { CLIENT_ID } from '@/constants'
import http from '@/service/http/mixin'

var crypto = require('crypto')

class OAuth {
  sha256(buffer) {
    return crypto.createHash('sha256').update(buffer).digest();
  }

  base64URLEncode(str) {
    return str.toString('base64')
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
  }

  requestCode() {
    const randomCode = crypto.randomBytes(32)
    const verifier = this.base64URLEncode(randomCode);
    const challenge = this.base64URLEncode(this.sha256(randomCode));
    localStorage.setItem('code-verifier', verifier)
    window.location.href = `https://mixin.one/oauth/authorize?client_id=${CLIENT_ID}&scope=PROFILE%3AREAD+ASSETS%3AREAD&code_challenge=${challenge}`
  }

  async getAccessToken(code) {
    const verifier = localStorage.getItem('code-verifier')
    const data = { client_id: CLIENT_ID, code, code_verifier: verifier }
    let res = await http.post('/oauth/token', { data });
    const { access_token } = res;
    localStorage.setItem('token', access_token);
  }
}

export default new OAuth()