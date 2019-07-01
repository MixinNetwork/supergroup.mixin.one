<template>
  <loading :loading="maskLoading" :fullscreen="true">
    <div class="coupons-page">
      <nav-bar :title="$t('coupons.title')"
        :hasTopRight="true" :hasBack="true"
        :rightText="$t('coupons.add_label')"
        @click-right="onClickRight"
        ></nav-bar>
      <van-list
        v-model="loading"
        :finished="finished"
        finished-text="~ END ~"
        @load="onLoad"
      >
        <coupon-item :coupon="item" v-for="item in items" @coupon-click="couponClick"></coupon-item>
      </van-list>
      <van-action-sheet
        v-model="showActionSheet"
        :actions="actions"
        :cancel-text="$t('comm.cancel')"
        @select="onSelectAction"
        @cancel="onCancelAction"
      />
      <van-dialog
        v-model="showAddCouponModel"
        :title="$t('coupons.add_model_title')"
        :message="$t('coupons.add_model_desc')"
        :confirm-button-text="$t('coupons.create_button_label')"
        :show-confirm-button="true"
        :show-cancel-button="true"
        @confirm="onCreateCoupons"
        @cancel="() => showAddCouponModel = false"
      ></van-dialog>
    </div>
  </loading>
</template>

<script>
import NavBar from '@/components/Nav'
import dayjs from 'dayjs'
import CouponItem from '@/components/partial/CouponItem'
import Loading from '@/components/Loading'
import { ActionSheet, Toast, Dialog } from 'vant'
import utils from '@/utils'
import { saveAs } from 'file-saver'

export default {
  name: 'Coupons',
  props: {
  },
  data () {
    return {
      showActionSheet: false,
      showAddCouponModel: false,
      maskLoading: false,
      currentCoupon: null,
      loading: false,
      finished: false,
      items: [],
      actions: [
        { name: this.$t('coupons.copy') },
      ]
    }
  },
  components: {
    NavBar, CouponItem, Loading,
    'van-action-sheet': ActionSheet,
  },
  async mounted () {
  },
  computed: {

  },
  methods: {
    async onLoad() {
      this.loading = false
      this.maskLoading = false
    },
    couponClick (mem) {
      // this.showActionSheet = true
    },
    async onSelectAction (item, ix) {
      if (this.currentCoupon) {
        let mem = this.currentCoupon
      }
      this.showActionSheet = false
    },
    onCancelAction (item, ix) {
      this.showActionSheet = false
    },
    onClickRight () {
      if (window.localStorage.getItem('role') === 'admin') {
        this.showAddCouponModel = true
      }
    },
    onCreateCoupons () {
      this.maskLoading = true
      this.GLOBAL.api.coupon.create().then((resp) => {
        const blob = new Blob([resp], {type: 'text/csv'})
        saveAs(blob, 'coupons.csv')
        //utils.reloadPage()
        this.maskLoading = false
      }).catch((err) => {
        console.log(err)
      })
    }
  }
}
</script>

<style scoped>
.coupons-page {
  padding-top: 60px;
}
</style>
