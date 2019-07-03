<template>
  <loading :loading="maskLoading" :fullscreen="true">
    <div class="messages-page">
      <nav-bar :title="$t('messages.title')" :hasTopRight="false" :hasBack="true"></nav-bar>
      <van-list
        v-model="loading"
        :finished="finished"
        finished-text="~ END ~"
        @load="onLoad"
      >
        <message-item :message="item" v-for="item in items" @message-click="messageClick"></message-item>
      </van-list>
      <van-action-sheet
        :title="currentMessage ? currentMessage.full_name : ''"
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
import MessageItem from '@/components/partial/MessageItem'
import Loading from '@/components/Loading'
import { ActionSheet, Toast } from 'vant'
import utils from '@/utils'

export default {
  name: 'Messages',
  props: {
  },
  data () {
    return {
      showActionSheet: false,
      maskLoading: false,
      currentMessage: null,
      loading: false,
      finished: false,
      items: [],
      actions: [
        { name: this.$t('messages.recall') },
      ]
    }
  },
  components: {
    NavBar, MessageItem, Loading,
    'van-action-sheet': ActionSheet,
  },
  async mounted () {
  },
  computed: {

  },
  methods: {
    async onLoad() {
      this.loading = true
      this.maskLoading = true
      let resp = await this.GLOBAL.api.message.index()
      this.loading = false
      this.maskLoading = false
      this.finished = true
      console.log(resp.data)
      this.items = resp.data.map((x) => {
        x.full_name = x.full_name === 'NULL' ? 'SYSTEM' : x.full_name
        x.time = dayjs(x.created_at).format('YYYY.MM.DD')
        return x
      })
    },
    messageClick (mem) {
      if (window.localStorage.getItem('role') === 'admin') {
        this.currentMessage = mem
        this.showActionSheet = true
      }
    },
    async onSelectAction (item, ix) {
      if (this.currentMessage) {
        this.maskLoading = true
        let mem = this.currentMessage
        if (ix === 0) {
          let result = await this.GLOBAL.api.message.recall(mem.message_id)
          if (result.error) {
            this.maskLoading = false
            return
          }
          utils.reloadPage()
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
.messages-page {
  padding-top: 60px;
}
</style>
