<template>
  <loading :loading="loading" :fullscreen="true">
  <div class="wxpay-page" style="padding-top: 60px;">
    <nav-bar :title="$t('wxpay.title')" :hasTopRight="false"></nav-bar>
    <van-panel :title="$t('wxpay.welcome')" :desc="$t('wxpay.welcome_desc')">
    </van-panel>
    <br/>
    <van-panel>
      <van-cell
        :title="$t('wxpay.price_label', {price: wechatPaymentAmount, unit: $t('currency.cny')})"
        >
      </van-cell>
      <div slot="footer">
        <van-cell>
          <van-button style="width: 100%" type="primary" :disabled="loading" @click="payWechat">{{$t('wxpay.pay_wechat')}}</van-button>
        </van-cell>
      </div>
    </van-panel>
  </div>
  </loading>
</template>

<script>
import NavBar from '@/components/Nav'
import Loading from '@/components/Loading'
import Row from '@/components/Nav'
import {Toast} from 'vant'
import { CLIENT_ID } from '@/constants'

export default {
  name: 'WxPay',
  props: {
    msg: String
  },
  data () {
    return {
      loading: false,
      accessToken: '',
      openId: '',
      userId: '',
      wechatPaymentAmount: '100',
    }
  },
  components: {
    NavBar, Loading
  },
  async mounted () {
    this.loading = true
    this.accessToken = this.$route.query.access_token
    this.openId = this.$route.query.open_id
    this.userId = this.$route.query.user_id
    console.log(this.accessToken, this.openId)
    let config = await this.GLOBAL.api.website.config()
    this.wechatPaymentAmount = config.data.wechat_payment_amount
    if (this.accessToken && this.openId) {
      if (typeof WeixinJSBridge == "undefined"){
        if( document.addEventListener ){
          document.addEventListener('WeixinJSBridgeReady', this.onBridgeReady, false);
        }else if (document.attachEvent){
          document.attachEvent('WeixinJSBridgeReady', this.onBridgeReady); 
          document.attachEvent('onWeixinJSBridgeReady', this.onBridgeReady);
        }
        alert("Please open this page in WeChat")
      } else {
        this.onBridgeReady();
      }
    } else {
      alert("Incorrect page params")
    }
  },
  computed: {
  },
  methods: {
    async payWechat () {
      this.loading = true
      let orderInfo = await this.GLOBAL.api.account.create_wx_pay({
        'open_id': this.openId,
        'user_id': this.userId
      })
      let data = orderInfo.data
      if (data 
        && data.order.OrderId 
        && data.pay_params.return_code !== 'FAIL'
        && data.pay_params.result_code !== 'FAIL'
        && data.pay_js_params) {
        this.loading = false
        this.invokePayment(data.order, data.pay_js_params)
      } else {
        this.loading = false
        let msg = data.pay_params.err_code_des || data.pay_params.return_msg || 'Unknown Error'
        Toast('Error: ' + msg)
      }
      // this.meInfo = await this.GLOBAL.api.account.check_wx_pay('f64c815e-c98b-45fc-978e-ae6620bdfce0')
    },
    onBridgeReady () {
      this.loading = false
      console.log('WeixinJSBridge Loaded:', WeixinJSBridge)
    },
    invokePayment (order, jsparams) {
      this.loading = true
      WeixinJSBridge.invoke(
        'getBrandWCPayRequest', {
          "appId": jsparams.appId, 
          "timeStamp": jsparams.timeStamp,
          "nonceStr": jsparams.nonceStr,
          "package": jsparams.package,
          "signType": jsparams.signType,
          "paySign": jsparams.paySign
        },
        (res) => {
          console.log(res)
          // alert(JSON.stringify(res))
          if (res.err_msg == "get_brand_wcpay_request:ok") {
            Toast(this.$t('wxpay.succ_toast'))
            this.$router.push('/wxpay/done')
          } else if (res.err_msg == "get_brand_wcpay_request:fail"){
            Toast(this.$t('wxpay.error_toast', {msg: res.err_msg}))
          }
          this.loading = false
        }
      );
    }
  }
}
</script>

<style scoped>
.wxpay-page {
  padding-top: 60px;
}
</style>
