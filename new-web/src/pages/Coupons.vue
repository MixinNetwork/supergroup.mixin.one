<template>
  <loading :loading="maskLoading" :fullscreen="true">
    <div class="coupons-page">
      <nav-bar :title="$t('coupons.title')"
        :hasTopRight="false" :hasBack="true"
        ></nav-bar>
      <div class="button-wrapper">
        <van-button style="width: 100%" type="primary" @click="onCreateCoupons">
          {{$t('coupons.add_label')}}
        </van-button>
      </div>
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
      // this.loading = true
      // this.maskLoading = true
      // let resp = await this.GLOBAL.api.coupon.index()
      // this.maskLoading = false
      // console.log(resp.data)
      // this.items = resp.data.map((x) => {
        //   x.time = dayjs(x.created_at).format('YYYY.MM.DD')
      //   return x
      // })
      this.loading = false
      this.finished = true
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
    onCreateCoupons () {
      if (window.localStorage.getItem('role') === 'admin') {
        this.showAddCouponModel = true
      }
    },
    onCreateCoupons () {
      this.maskLoading = true
      this.GLOBAL.api.coupon.create({quantity: 100}).then((resp) => {
        const blob = new Blob([resp], {type: 'text/csv'})
        saveAs(blob, 'coupons.csv')
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
.button-wrapper {
  padding: 40px;
}
</style>
