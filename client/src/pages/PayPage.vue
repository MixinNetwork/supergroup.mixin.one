<template>
  <loading :loading="loading" :fullscreen="true">
  <div class="container">
    <nav-bar :title="$t('pay.title')"></nav-bar>
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
        <span>{{selectedAsset.symbol}}</span>
      </row-select>
      <van-cell
        :title="$t('pay.price_label', {price: selectedAsset.amount, unit: selectedAsset.symbol})"
        >
      </van-cell>
      <div>
        <van-cell>
          <van-button style="width: 100%" type="primary" :disabled="selectedAsset.amount==0 || loading" @click="payCrypto">{{$t('pay.pay_crypto')}}</van-button>
        </van-cell>
      </div>
    </van-panel>
    <div class='notice'>
    {{$t('pay.notice')}}
    </div>
  </div>
  </loading>
</template>

<script>
import NavBar from '@/components/NavBar'
import RowSelect from '@/components/RowSelect'
import Loading from '../components/LoadingSpinner'
import {Toast} from 'vant'
import { CLIENT_ID } from '@/constants'

export default {
  name: 'PayPage',
  props: {
    msg: String
  },
  data () {
    return {
      loading: false,
      config: null,
      meInfo: null,
      selectedAsset: {
        asset_id: "",
        symbol: "Tap To Select",
        amount: 0,
      },
      assets: []
    }
  },
  components: {
    NavBar, RowSelect, Loading
  },
  async mounted () {
    this.loading = true;
    let config = await this.GLOBAL.api.website.config();
    console.log(config);
    this.assets = config.data.accept_asset_list.map((a) => {
      a.name = a.symbol;
      a.amount = Math.floor(parseFloat(a.amount) * 100000000) / 100000000;
      return a
    });
    if (this.assets.length > 0) {
      this.selectedAsset = this.assets[0]
    }
    this.meInfo = await this.GLOBAL.api.account.me()
    this.loading = false
  },
  computed: {
  },
  methods: {
    payCrypto () {
      this.loading = true
      let traceId = this.meInfo.data.trace_id
      setTimeout(async () => { await this.waitForPayment(); }, 1000)
      window.location.href = `mixin://pay?recipient=${CLIENT_ID}&asset=${this.selectedAsset.asset_id}&amount=${this.selectedAsset.amount}&trace=${traceId}&memo=PAY_TO_JOIN`
    },
    async onChangeAsset (ix) {
      this.selectedAsset = this.assets[ix];
    },
    async waitForPayment () {
      let meInfo = await this.GLOBAL.api.account.me()
      if (meInfo.error) {
        setTimeout(async () => { await this.waitForPayment(); }, 1500)
        return;
      }
      if (meInfo.data.state === 'paid') {
        Toast(this.$t('pay.success_toast'))
        this.$router.push('/');
        this.loading = false
        return;
      }
      setTimeout(async () => { await this.waitForPayment(); }, 1500)
    },
  }
}
</script>

<style scoped>
.container {
  padding-top: 60px;
}
.notice {
  margin-top: 1.6rem;
}
</style>
