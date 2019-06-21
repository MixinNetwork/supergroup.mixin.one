<template>
  <loading :loading="loading" :fullscreen="true">
    <div class="home home-page page">
      <van-panel :title="$t('home.welcome')" :desc="$t('home.welcome_desc', {count: websiteInfo ? websiteInfo.data.users_count : '...'})">
      </van-panel>
      <br/>
      <template v-for="group in shortcutsGroups">
        <van-panel :title="group.label">
          <cell-table :items="group.shortcuts"></cell-table>
        </van-panel>
        <br/>
      </template>
      <van-panel :title="$t('home.pane_operations')" >
        <cell-table :items="builtinItems"></cell-table>
      </van-panel>
    </div>
  </loading>
</template>

<script>
import NavBar from '../components/Nav'
import CellTable from '../components/CellTable'
import Loading from '../components/Loading'
import { mapState } from 'vuex'
import AssetItem from '@/components/partial/AssetItem'
 
export default {
  data () {
    return {
      loading: false,
      meInfo: null,
      websiteInfo: null,
      builtinItems: [
        // builtin
        { icon: require('../assets/images/luckymoney-circle.png'), label: this.$t('home.op_luckycoin'), url: '/#/packets/prepare' },
        { icon: require('../assets/images/users-circle.png'), label: this.$t('home.op_members'), url: '/#/members' },
        { icon: require('../assets/images/messages-circle.png'), label:  this.$t('home.op_messages'), url: '/#/messages' },
        // { icon: require('../assets/images/notification-circle.png'), label: 'Subscribe', url: '' },
        // { icon: require('../assets/images/notification-off-circle.png'), label: 'Unsubscribe', url: '' },
      ],
      shortcutsGroups: []
    }
  },
  computed: {
    isSubscribed () {
      if (this.meInfo) {
        if (new Date(this.meInfo.data.subscribed_at).getYear() < 0) {
          return false
        }
      }
      return true
    },
    isProhibited () {
      return this.websiteInfo && this.websiteInfo.data.prohibited
    },
    isZh() {
      return this.$i18n.locale.indexOf('zh') !== -1
    }
  },
  components: {
    NavBar, CellTable, Loading
  },
  async mounted () {
    try {
      this.loading = true
      this.GLOBAL.api.website.config().then((conf) => {
        this.shortcutsGroups = conf.data.home_shortcut_groups.map((x) => {
          x.label = this.isZh ? x.label_zh: x.label_en
          x.shortcuts = x.shortcuts.map((z) => {
            z.label = this.isZh ? z.label_zh: z.label_en
            return z
          })
          return x
        })
        this.loading = false
      })
      this.websiteInfo = await this.GLOBAL.api.website.amount()
      this.meInfo = await this.GLOBAL.api.account.me()
      if (this.meInfo.data.state === 'pending') {
        this.$router.replace('/pay')
        return
      }
      if (this.meInfo.data.role === 'admin') {
        this.updateProhibitedState()
      }
      this.updateSubscribeState()
    } catch (err) {
      console.log('error', err)
    }
  },
  methods: {
    updateSubscribeState() {
      if (!this.isSubscribed) {
        this.builtinItems.push({
          icon: require('../assets/images/notification-circle.png'), label: this.$t('home.op_subscribe'),
          click: async (evt) => {
            evt.preventDefault()
            await this.GLOBAL.api.account.subscribe()
            window.location.reload()
          }
        })
      } else {
        this.builtinItems.push({
          icon: require('../assets/images/notification-off-circle.png'), label: this.$t('home.op_unsubscribe'),
          click: async (evt) => {
            evt.preventDefault()
            await this.GLOBAL.api.account.unsubscribe()
            window.location.reload()
          }
        })
      }
    },
    updateProhibitedState() {
      if (!this.isProhibited) {
        this.builtinItems.push({
          icon: require('../assets/images/unprohibited.png'), label: this.$t('home.op_mute'),
          click: async (evt) => {
            evt.preventDefault()
            await this.GLOBAL.api.property.create(true)
            window.location.reload()
          }
        })
      } else {
        this.builtinItems.push({
          icon: require('../assets/images/prohibited.png'), label: this.$t('home.op_unmute'),
          click: async (evt) => {
            evt.preventDefault()
            await this.GLOBAL.api.property.create(false)
            window.location.reload()
          }
        })
      }
    }
  }
}
</script>

<style lang="scss" scoped>
.home-page {
  // padding-top: 60px;
}
</style>

