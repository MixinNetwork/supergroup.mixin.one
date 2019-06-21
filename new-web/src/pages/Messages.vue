<template>
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
</template>

<script>
import NavBar from '@/components/Nav'
import dayjs from 'dayjs'
import MessageItem from '@/components/partial/MessageItem'
import { ActionSheet, Toast } from 'vant'

export default {
  name: 'Messages',
  props: {
  },
  data () {
    return {
      showActionSheet: false,
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
    NavBar, MessageItem, 
    'van-action-sheet': ActionSheet,
  },
  async mounted () {
  },
  computed: {

  },
  methods: {
    async onLoad() {
      this.loading = true
      let resp = await this.GLOBAL.api.message.index()
      this.loading = false
      this.finished = true
      console.log(resp.data)
      this.items = resp.data.map((x) => {
        x.full_name = x.full_name === 'NULL' ? 'SYSTEM' : x.full_name
        x.time = dayjs(x.created_at).format('YYYY.MM.DD')
        return x
      })
    },
    messageClick (mem) {
      this.currentMessage = mem
      this.showActionSheet = true
    },
    async onSelectAction (item, ix) {
      if (this.currentMessage) {
        let mem = this.currentMessage
        if (ix === 0) {
          Toast(this.$t('messages.recall') + '...')
          let result = await this.GLOBAL.api.message.recall(mem.message_id)
          if (result.error) {
            Toast(result.error.toString())
            return
          }
          window.location.reload()
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
