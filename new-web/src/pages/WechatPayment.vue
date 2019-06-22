<template>
  <loading :loading="maskLoading" :fullscreen="true">
    <div class="wxpay-page">
      <nav-bar :title="$t('wxpay.title')" :hasTopRight="false" :hasBack="true"></nav-bar>
      <van-panel :title="$t('wxpay.hint_title')" :desc="$t('wxpay.hint_desc')">
        <van-row>
          <vue-qr style="display: block; margin: 0 auto" :text="qrUrl" :size="200"></vue-qr>
        </van-row>
      </van-panel>
      <br/>
      <van-panel :title="$t('wxpay.help_title')">
        <van-row>
          <ol class="help-list">
            <li>{{$t('wxpay.help_line_1')}}
            </li>
            <li>{{$t('wxpay.help_line_2')}}
            </li>
            <li>{{$t('wxpay.help_line_3')}}
              <img :src="require('../assets/images/wxpay/step_3_zh.jpg')"/>
            </li>
            <li>{{$t('wxpay.help_line_4')}}
              <img :src="require('../assets/images/wxpay/step_4_zh.jpg')"/>
            </li>
          </ol>
        </van-row>
      </van-panel>
    </div>
  </loading>
</template>

<script>
import NavBar from '@/components/Nav'
import dayjs from 'dayjs'
import Loading from '@/components/Loading'
import { ActionSheet, Toast } from 'vant'
import utils from '@/utils'

export default {
  name: 'WxPay',
  data () {
    return {
      maskLoading: false,
      qrUrl: '',
      orderId: '',
    }
  },
  components: {
    NavBar, Loading
  },
  async mounted () {
    this.orderId = this.$route.params.id
    this.qrUrl = this.$route.query.qr_url
    setTimeout(() => {
      this.waitForPayment()
    }, 5000)
  },
  computed: {

  },
  methods: {
    async waitForPayment() {
      if (this.orderId) {
        let orderInfo = await this.GLOBAL.api.account.check_wx_pay(this.orderId)
        console.log(orderInfo)
        if (orderInfo.data && orderInfo.data.state === 'PAID') {
          this.$router.push('/')
          return
        } else {
          setTimeout(() => {
            this.waitForPayment()
          }, 30000)
        }
      } else {
        setTimeout(() => {
          this.waitForPayment()
        }, 30000)
      }
    }
  }
}
</script>

<style scoped>
.wxpay-page {
  padding-top: 60px;
}
.help-list {
  padding: 16px 20px;
}
.help-list li {
  margin-bottom: 10px;
}
.help-list li img {
  width: 60%;
  border-radius: 5px;
  display: block;
  margin: 10px auto;
}
</style>
