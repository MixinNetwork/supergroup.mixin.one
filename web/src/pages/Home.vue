<template>
  <loading :loading="loading" :fullscreen="true">
    <div class="home home-page page">
      <van-panel :title="welcomeMessage || $t('home.welcome')" :desc="$t('home.welcome_desc', {count: websiteInfo ? websiteInfo.data.users_count : '...'})">
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
import utils from '@/utils'

export default {
  data () {
    return {
      loading: false,
      meInfo: null,
      welcomeMessage: '',
      websiteInfo: null,
      websiteConf: null,
      builtinItems: [
        // builtin
        { icon: require('../assets/images/luckymoney-circle.png'), label: this.$t('home.op_luckycoin'), url: '/packets/prepare' },
        { icon: require('../assets/images/users-circle.png'), label: this.$t('home.op_members'), url: '/members' },
      ],
      messagesItem: {
        icon: require('../assets/images/messages-circle.png'), label:  this.$t('home.op_messages'), url: '/messages'
      },
      couponsItem: {
        icon: require('../assets/images/coupons.png'), label:  this.$t('home.op_coupons'), url: '/coupons'
      },
      // 订阅始终在倒数第一个位置
      subscribeItem: {
        icon: require('../assets/images/notification-circle.png'), label: this.$t('home.op_subscribe'),
        click: async (evt) => {
          evt.preventDefault()
          await this.GLOBAL.api.account.subscribe()
          this.builtinItems.splice(this.builtinItems.length - 1, 1, this.unsubscribeItem)

        }
      },
      unsubscribeItem: {
        icon: require('../assets/images/notification-off-circle.png'), label: this.$t('home.op_unsubscribe'),
        click: async (evt) => {
          evt.preventDefault()
          await this.GLOBAL.api.account.unsubscribe()
          this.builtinItems.splice(this.builtinItems.length - 1, 1, this.subscribeItem)
        }
      },
      // 禁言始终在倒数第二个位置
      unprohibitItem: {
        icon: require('../assets/images/unprohibited.png'), label: this.$t('home.op_unmute'),
        click: async (evt) => {
          evt.preventDefault()
          await this.GLOBAL.api.property.create(false)
          this.builtinItems.splice(this.builtinItems.length - 2, 1, this.prohibitItem)
        }
      },
      prohibitItem: {
        icon: require('../assets/images/prohibited.png'), label: this.$t('home.op_mute'),
        click: async (evt) => {
          evt.preventDefault()
          await this.GLOBAL.api.property.create(true)
          this.builtinItems.splice(this.builtinItems.length - 2, 1, this.unprohibitItem)
        }
      },
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
        this.websiteConf = conf
        if (conf.data.home_shortcut_groups) {
          this.shortcutsGroups = this.addToGroups(conf.data.home_shortcut_groups, false)
        }
        this.GLOBAL.api.plugin.shortcuts().then((resp) => {
          this.shortcutsGroups = this.addToGroups(resp.data, true)
        })
        this.welcomeMessage = this.websiteConf.data.home_welcome_message
        this.loading = false
      })

      this.websiteInfo = await this.GLOBAL.api.website.amount()
      this.meInfo = await this.GLOBAL.api.account.me()
      if (this.meInfo.data.state === 'pending') {
        this.$router.push('/pay')
        return
      }
      if (this.meInfo.data.role === 'admin') {
        this.builtinItems.push(this.couponsItem)
        this.builtinItems.push(this.messagesItem)
        this.updateProhibitedState()
      }
      this.updateSubscribeState()
    } catch (err) {
      console.log('error', err)
    }
  },
  methods: {
    addToGroups (groups, isPlugin) {
      return this.shortcutsGroups.concat(groups.map((x) => {
        x.label = this.isZh ? x.label_zh: x.label_en
        const items = x.items || x.shortcuts
        x.shortcuts = items.map((z) => {
          z.label = this.isZh ? z.label_zh: z.label_en
          if (isPlugin) {
            // z.click = this.handlePluginRedirect(x.id, z.id)
            // z.url = ''
            z.url += '?token=' + encodeURIComponent(localStorage.getItem('token'))
          }
          return z
        })
        return x
      }))
    },
    handlePluginRedirect(groupId, itemId) {
      return () => {
        this.GLOBAL.api.plugin.redirect(groupId, itemId).then((resp) => {
          console.log(resp)
        }).catch((err) => {
          console.log(err)
        })
      }
    },
    updateSubscribeState() {
      if (this.isSubscribed) {
        this.builtinItems.push(this.unsubscribeItem)
      } else {
        this.builtinItems.push(this.subscribeItem)
      }
    },
    updateProhibitedState() {
      if (this.isProhibited) {
        this.builtinItems.push(this.unprohibitItem)
      } else {
        this.builtinItems.push(this.prohibitItem)
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

