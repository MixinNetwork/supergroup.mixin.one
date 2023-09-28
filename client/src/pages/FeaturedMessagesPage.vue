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
        <featured-message-item :message="item" v-for="item in items"  v-bind:key="item.message_id" @message-click="messageClick"></featured-message-item>
      </van-list>
    </div>
  </loading>
</template>

<script>
import NavBar from '@/components/NavBar'
import dayjs from 'dayjs'
import FeaturedMessageItem from '@/components/partial/FeaturedMessageItem'
import Loading from '@/components/LoadingSpinner'
import utils from '@/utils'

export default {
  name: 'FeaturedMessagesPage',
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
    NavBar, FeaturedMessageItem, Loading,
  },
  async mounted () {
  },
  computed: {

  },
  methods: {
    async onLoad() {
      this.loading = true
      this.maskLoading = true
      let resp = await this.GLOBAL.api.message.featured_messages()
      this.loading = false
      this.maskLoading = false
      this.finished = true
      this.items = resp.data.map((x) => {
        x.full_name = x.full_name === 'NULL' ? 'SYSTEM' : x.full_name
        x.time = dayjs(x.created_at).format('YYYY.MM.DD')
        return x
      })
    },
    messageClick (mem) {
      if (window.localStorage.getItem('role') === 'admin') {
        this.currentMessage = mem
        this.showActionSheet = false
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
    onCancelAction () {
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
