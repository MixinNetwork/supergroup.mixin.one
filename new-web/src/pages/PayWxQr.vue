<template>
  <div class="wxpay-qr-page">
    <nav-bar :title="$t('pay_wxqr.title')" :hasTopRight="false" :hasBack="true"></nav-bar>
    <van-panel :title="$t('pay_wxqr.hint_title')" :desc="$t('pay_wxqr.hint_desc')">
      <van-row>
        <vue-qr style="display: block; margin: 0 auto" :text="qrUrl" :size="200"></vue-qr>
      </van-row>
    </van-panel>
    <br/>
    <van-panel :title="$t('pay_wxqr.help_title')">
      <van-row>
        <ol class="help-list">
          <li>{{$t('pay_wxqr.help_line_1')}}
          </li>
          <li>{{$t('pay_wxqr.help_line_2')}}
          </li>
          <li>{{$t('pay_wxqr.help_line_3')}}
            <img :src="require('../assets/images/wxpay/step_3_zh.jpg')"/>
          </li>
          <li>{{$t('pay_wxqr.help_line_4')}}
            <img :src="require('../assets/images/wxpay/step_4_zh.jpg')"/>
          </li>
        </ol>
      </van-row>
    </van-panel>
  </div>
</template>

<script>
import NavBar from '@/components/Nav'
import dayjs from 'dayjs'
import { ActionSheet, Toast } from 'vant'
import utils from '@/utils'

export default {
  name: 'WxPayQr',
  data () {
    return {
      qrUrl: '',
    }
  },
  components: {
    NavBar
  },
  async mounted () {
    this.qrUrl = this.$route.query.qr_url
  },
  computed: {

  },
  methods: {
  }
}
</script>

<style scoped>
.wxpay-qr-page {
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
