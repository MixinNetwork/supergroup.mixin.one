<template>
  <div class="pay-page">
    <nav-bar :title="$t('pay.title')" :hasTopRight="false"></nav-bar>
    <van-panel :title="$t('pay.welcome')" :desc="$t('pay.welcome_desc')">
    </van-panel>
    <br/>
    <van-panel :title="$t('pay.method_crypto')">
      <row-select 
        :index="0"
        :title="$t('pay.select_assets')"
        :columns="assets"
        placeholder="Tap to Select"
        @change="onChangeAsset">
        <span slot="text">{{selectedAsset ? selectedAsset.text : 'Tap to Select'}}</span>
      </row-select>
      <van-cell
        :title="$t('pay.price_label', {price: selectedAsset ? selectedAsset.price: '...', unit: selectedAsset ? selectedAsset.text : '...'})"
        >
        <span>约 ¥19.9</span>
      </van-cell>
      <div slot="footer">
        <van-cell>
          <van-button style="width: 100%" type="info" :disabled="selectedAsset === null" @click="payCrypto">{{$t('pay.pay_crypto')}}</van-button>
        </van-cell>
        <van-cell>
          <van-button style="width: 100%" type="warning" :disabled="selectedAsset === null" @click="payCrypto">{{$t('pay.pay_foxone')}}</van-button>
        </van-cell>
      </div>
    </van-panel>
    <br/>
    <van-panel :title="$t('pay.method_wechat')">
      <van-cell
        :title="$t('pay.price_label', {price: '19.9', unit: '元'})"
        >
      </van-cell>
      <div slot="footer">
        <van-cell>
          <van-button style="width: 100%" type="primary" @click="payCrypto">{{$t('pay.pay_wechat')}}</van-button>
        </van-cell>
      </div>
    </van-panel>
  </div>
</template>

<script>
import NavBar from '@/components/Nav'
import RowSelect from '@/components/RowSelect'
import Row from '@/components/Nav'
import { CLIENT_ID } from '@/constants'
import uuid from 'uuid'
export default {
  name: 'Pay',
  props: {
    msg: String
  },
  data () {
    return {
      showPicker: false,
      selectedAsset: null,
      assets: [
        { text: 'XIN', assetId: 'c94ac88f-4671-3976-b60a-09064f1811e8', price: '0.0119' },
        { text: 'PRS', assetId: 'c94ac88f-4671-3976-b60a-09064f1811e8', price: '0.0119' },
      ]
    }
  },
  components: {
    NavBar, RowSelect
  },
  methods: {
    payCrypto () {
      setTimeout(() => { this.waitForPayment(); }, 2000)
      let trace_id = uuid.v4()
      window.location.replace(`mixin://pay?recipient=${CLIENT_ID}&asset=${this.selectedAsset.assetId}&amount=${this.selectedAsset.price}&trace=${trace_id}&memo=PAY_TO_JOIN`);
    },
    onChangeAsset (ix) {
      this.selectedAsset = this.assets[ix]
    },
    waitForPayment () {
      let meInfo = this.GLOBAL.api.account.me()
      if (meInfo.error) {
        setTimeout(() => { this.waitForPayment(); }, 1000)
        return;
      }
      if (meInfo.data.state === 'paid') {
        this.$router.replace('/');
        return;
      }
      setTimeout(() => { this.waitForPayment(); }, 1000)
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h3 {
  margin: 40px 0 0;
}
ul {
  list-style-type: none;
  padding: 0;
}
li {
  display: inline-block;
  margin: 0 10px;
}
a {
  color: #42b983;
}
</style>
