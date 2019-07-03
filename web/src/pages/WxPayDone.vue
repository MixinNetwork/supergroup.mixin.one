<template>
  <div class="wxpay-done-page" style="padding-top: 60px;">
    <nav-bar :title="$t('wxpay_done.title')"></nav-bar>
    <van-panel>
      <van-icon class="icon" name="success" size="48" />
      <van-row>
        <p class="hint" v-html="$t('wxpay_done.hint')"></p>
      </van-row>
    </van-panel>
  </div>
</template>

<script>
import NavBar from '@/components/Nav'
import Row from '@/components/Nav'
import { Icon } from 'vant';
export default {
  name: 'WxPayDone',
  data () {
    return {
    }
  },
  components: {
    NavBar,
    "van-icon": Icon
  },
  computed: {
  },
  methods: {
  }
}
</script>

<style scoped>
.wxpay-done-page {
  padding-top: 60px;
}
.icon {
  margin: 10px auto 40px auto;
  text-align: center;
  display: block;
  color: yellowgreen;
}
.hint {
  margin: 10px auto;
  text-align: center;
  margin-bottom: 40px;
}
</style>
