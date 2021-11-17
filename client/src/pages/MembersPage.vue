<template>
  <loading :loading="maskLoading" :fullscreen="true">
    <div class="members-page">
      <nav-bar :title="$t('members.title')" :hasTopRight="false" :hasBack="true"></nav-bar>
      <van-cell>
        <van-field placeholder="Search" left-icon="search"
          @change="searchEnter" v-model="searchQuery"
          >
        </van-field>
      </van-cell>
      <van-list
        v-model="loading"
        :finished="finished"
        finished-text="~ END ~"
        @load="onLoad"
      >
      <member-item :member="item" v-for="item in items" v-bind:key="item.full_name" @member-click="memberClick"></member-item>
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
import NavBar from '@/components/NavBar'
import dayjs from 'dayjs'
import MemberItem from '@/components/partial/MemberItem'
import Loading from '@/components/LoadingSpinner'
import { ActionSheet } from 'vant'
import utils from '@/utils'

export default {
  name: 'MembersPage',
  props: {
  },
  data () {
    return {
      searchQuery: '',
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
      await this.loadMembers(this.lastOffset, '')
    },
    async loadMembers(offset=0, query='', append=true) {
      this.maskLoading = true
      this.loading = true
      let role = window.localStorage.getItem('role')
      let resp = await this.GLOBAL.api.account.subscribers(offset, query)
      if (resp.data.length < 2) {
        this.finished = true
      }
      resp.data = resp.data.map((x) => {
        x.time = dayjs(x.subscribed_at).format('YYYY.MM.DD')
        if (role !== 'admin') {
          x.identity_number = '0'
        }
        return x
      })
      if (append) {
        this.items = this.items.concat(resp.data)
      } else {
        this.items = resp.data
        this.finished = true
      }
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
          let result = await this.GLOBAL.api.account.block(mem.user_id)
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
    onCancelAction () {
      this.showActionSheet = false
    },
    searchEnter () {
      this.loadMembers(0, this.searchQuery, false)
      this.finished = true
    }
  }
}
</script>

<style scoped>
.members-page {
  padding-top: 60px;
}
</style>
