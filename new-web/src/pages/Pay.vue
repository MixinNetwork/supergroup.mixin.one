<template>
  <div class="pay-page">
    <nav-bar :title="$t('pay.title')" :hasTopRight="false"></nav-bar>
    <van-panel :title="$t('pay.welcome')" :desc="$t('pay.welcome_desc')">
    </van-panel>
    <br/>
    <van-panel :title="$t('pay.method_crypto')">
      <van-cell
        :title="$t('pay.select_assets')"
        >
        <van-field
          readonly
          clickable
          :value="selectedAsset ? selectedAsset.text: ''"
          placeholder="Tap to Select"
          @click="showPicker = true"
        />
      </van-cell>
      <van-cell
        :title="$t('pay.price_label', {price: selectedAsset ? selectedAsset.price: '...', unit: selectedAsset ? selectedAsset.text : '...'})"
        >
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
      <div slot="footer">
        <van-cell>
          <van-button style="width: 100%" type="primary" @click="payCrypto">{{$t('pay.pay_wechat')}}</van-button>
        </van-cell>
      </div>
    </van-panel>
    <van-popup v-model="showPicker" position="bottom">
      <van-picker
        show-toolbar
        :columns="assets"
        :default-index="0"
        @cancel="showPicker = false"
        @change="onChangeAsset"
        @confirm="onConfirmAsset"
      />
    </van-popup>
  </div>
</template>

<script>
import NavBar from '@/components/Nav'
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
        { text: 'XIN', assetId: 'c94ac88f-4671-3976-b60a-09064f1811e8', price: '0.02' },
        { text: 'PRS', assetId: 'c94ac88f-4671-3976-b60a-09064f1811e8', price: '0.02' },
      ]
    }
  },
  components: {
    NavBar
  },
  methods: {
    payCrypto () {
      setTimeout(() => { this.waitForPayment(); }, 2000)
      let trace_id = uuid.v4()
      window.location.replace(`mixin://pay?recipient=${CLIENT_ID}&asset=${this.selectedAsset.assetId}&amount=${this.selectedAsset.price}&trace=${trace_id}&memo=PAY_TO_JOIN`);
    },
    onChangeAsset (picker, value, ix) {
    },
    onConfirmAsset (value, ix) {
      this.selectedAsset = this.assets[ix]
      this.showPicker = false
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
