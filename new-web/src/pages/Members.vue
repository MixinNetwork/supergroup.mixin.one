<template>
  <loading :loading="maskLoading" :fullscreen="true">
    <div class="members-page">
      <nav-bar :title="$t('members.title')" :hasTopRight="false" :hasBack="true"></nav-bar>
      <van-list
        v-model="loading"
        :finished="finished"
        finished-text="~ END ~"
        @load="onLoad"
      >
        <member-item :member="item" v-for="item in items" @member-click="memberClick"></member-item>
      </van-list>
      <van-action-sheet
        :title="currentMember ? currentMember.full_name : ''"
        v-model="showActionSheet"
        :actions="actions"
        :cancel-text="$t('comm.cancel')"
        @select="onSelectAction"
        @cancel="onCancelAction"
      />
    </div>
  </loading>
</template>

<script>
import NavBar from '@/components/Nav'
import dayjs from 'dayjs'
import MemberItem from '@/components/partial/MemberItem'
import Loading from '@/components/Loading'
import { ActionSheet, Toast } from 'vant'
import utils from '@/utils'

export default {
  name: 'Members',
  props: {
  },
  data () {
    return {
      showActionSheet: false,
      maskLoading: false,
      currentMember: null,
      loading: false,
      finished: false,
      items: [],
      actions: [
        { name: this.$t('members.kick') },
        { name: this.$t('members.block') },
      ]
    }
  },
  components: {
    NavBar, MemberItem, Loading,
    'van-action-sheet': ActionSheet,
  },
  async mounted () {
  },
  computed: {
    lastOffset () {
      if (this.items.length) {
        let d = new Date(this.items[this.items.length - 1].subscribed_at)
        d.setSeconds(d.getSeconds() + 1)
        return d.toISOString()
      }
      return 0
    }
  },
  methods: {
    async onLoad() {
      this.maskLoading = true
      this.loading = true
      let resp = await this.GLOBAL.api.account.subscribers(this.lastOffset)
      if (resp.data.length < 2) {
        this.finished = true
      }
      resp.data = resp.data.map((x) => {
        x.time = dayjs(x.subscribed_at).format('YYYY.MM.DD')
        return x
      })
      this.items = this.items.concat(resp.data)
      this.loading = false
      this.maskLoading = false
    },
    memberClick (mem) {
      if (window.localStorage.getItem('role') === 'admin') {
        this.currentMember = mem
        this.showActionSheet = true
      }
    },
    async onSelectAction (item, ix) {
      if (this.currentMember) {
        let mem = this.currentMember
        this.maskLoading = true
        if (ix === 0) {
          let result = await this.GLOBAL.api.account.remove(mem.user_id)
          if (result.error) {
            this.maskLoading = false
            return
          }
          utils.reloadPage()
        } else if (ix === 1) {
          let result = await this.GLOBAL.api.account.remove(mem.user_id)
          if (result.error) {
            this.maskLoading = false
            return
          }
          utils.reloadPage()
        } else {
          this.maskLoading = false
        }
      }
      this.showActionSheet = false
    },
    onCancelAction (item, ix) {
      this.showActionSheet = false
    },
  }
}
</script>

<style scoped>
.members-page {
  padding-top: 60px;
}
</style>
