<template>
  <div class="home home-page page">
    <nav-bar :title="$t('home.title')" :hasTopRight="false"></nav-bar>
    <van-panel :title="$t('home.welcome')" :desc="$t('home.welcome_desc', {count: websiteInfo ? websiteInfo.data.users_count : '...'})">
    </van-panel>
    <br/>
    <van-panel :title="$t('home.pane_operations')" >
      <cell-table title="Built-in" :items="builtinItems"></cell-table>
    </van-panel>
    <br/>
    <van-panel :title="$t('home.pane_features')" >
      <cell-table title="Community" :items="CommunityItems"></cell-table>
    </van-panel>
    </div>
  </div>
</template>

<script>
import NavBar from '../components/Nav'
import CellTable from '../components/CellTable'
import { mapState } from 'vuex'
import AssetItem from '@/components/partial/AssetItem'
 
export default {
  data () {
    return {
      meInfo: null,
      websiteInfo: null,
      builtinItems: [
        // builtin
        // { icon: require('../assets/images/notification-circle.png'), label: 'Subscribe', url: '' },
        // { icon: require('../assets/images/notification-off-circle.png'), label: 'Unsubscribe', url: '' },
        { icon: require('../assets/images/users-circle.png'), label: 'Members', url: '/members' },
      ],
      CommunityItems: [
        { icon: require('../assets/images/luckymoney-circle.png'), label: 'Lucky Coin', url: '/luckycoin' },
        { icon: require('../assets/images/gbi.news.png'), label: 'GBI.news', url: 'https://gbi.news' },
        { icon: require('../assets/images/wallet-circle.png'), label: 'Wallet', url: 'https://elite-wallet.kumiclub.com' },
      ]
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
    }
  },
  components: {
    NavBar, CellTable
  },
  async mounted () {
    try {
      this.websiteInfo = await this.GLOBAL.api.website.amount()
      this.meInfo = await this.GLOBAL.api.account.me()
      console.log(this.meInfo)
      if (this.meInfo.data.state !== 'pending') {
        this.$router.replace('/pay')
        return
      }
      if (!this.isSubscribed) {
        this.builtinItems.unshift({
          icon: require('../assets/images/notification-circle.png'), label: 'Subscribe',
          click: async (evt) => {
            evt.preventDefault()
            await this.GLOBAL.api.account.subscribe()
            window.location.reload()
          }
        })
      } else {
        this.builtinItems.unshift({
          icon: require('../assets/images/notification-off-circle.png'), label: 'Unsubscribe',
          click: async (evt) => {
            evt.preventDefault()
            await this.GLOBAL.api.account.unsubscribe()
            window.location.reload()
          }
        })
      }
    } catch (err) {
      console.log('error', err)
    }
  },
  methods: {
  }
}
</script>

<style lang="scss" scoped>
.table-wrapper {
  // padding: 0 10px;
}
</style>

