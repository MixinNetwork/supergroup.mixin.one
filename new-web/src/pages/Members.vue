<template>
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
</template>

<script>
import NavBar from '@/components/Nav'
import dayjs from 'dayjs'
import MemberItem from '@/components/partial/MemberItem'
import { ActionSheet, Toast } from 'vant'

export default {
  name: 'Members',
  props: {
  },
  data () {
    return {
      showActionSheet: false,
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
    NavBar, MemberItem, 
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
    },
    memberClick (mem) {
      this.currentMember = mem
      this.showActionSheet = true
    },
    async onSelectAction (item, ix) {
      if (this.currentMember) {
        let mem = this.currentMember
        if (ix === 0) {
          Toast(this.$t('members.kick') + mem.full_name)
          let result = await this.GLOBAL.api.account.remove(mem.user_id)
          if (result.error) {
            Toast(result.error.toString())
            return
          }
        } else if (ix === 1) {
          Toast(this.$t('members.kick') + mem.full_name)
          let result = await this.GLOBAL.api.account.remove(mem.user_id)
          if (result.error) {
            Toast(result.error.toString())
            return
          }
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
