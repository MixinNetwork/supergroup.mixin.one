<template>
  <loading :loading="loading" :fullscreen="true">
  <div class="packet-page" :class="isClose ? '' : 'open'">
    <div class="packet header">
      <div class="user avatar">
        <img v-if="hasAvatar" :src="user ? user.avatar_url : '#'" alt="user avatar"/>
        <p v-else>{{firstLetter}}</p>
      </div>
      <h1 class="user name">{{user ? user.full_name : '...'}}</h1>
      <h2 class="greeting" v-if="pktData">
        <i class="icon-bubble"></i> 
        {{pktData.greeting || '...'}}
      </h2>
    </div>

    <div v-if="!isClose" class="packet lottery">
      <template v-if="lottery">
        <h3>{{lottery.amount}}<span>{{asset.symbol}}</span></h3>
        <p>{{$t('packet.paid')}}</p>
      </template>
      <template v-else>
        <h3 class="expire statement">{{$t('packet.completed')}}</h3>
      </template>
    </div>
    <div v-else class="packet open button">
      <button @click="openPacket">{{$t('packet.open')}}</button>
    </div>

    <template v-if="pktData">
      <div v-if="!isClose" class="packet history">
        <h4>{{$t('packet.opened', {opened_count: pktData.opened_count, total_count: pktData.total_count})}}, 
          {{pktData.opened_amount}}/{{pktData.amount}} {{asset.symbol}}</h4>
        <ul >
          <li v-for="user in pktData.participants">
            <div class="user avatar">
              <p v-if="user.firstLetter">{{user.firstLetter}}</p>
              <img v-else :src="user.avatar_url" alt="user avatar"/>
            </div>
            <div class="user info">
              <h5>{{user.full_name}}</h5>
              <time>{{user.created_at}}</time>
            </div>
            <div class="packet amount">{{user.amount}} {{user.symbol}}</div>
          </li>
        </ul>
      </div>
      <div class="submitting overlay">
        <div class="spinner-container">
          <div class="spinner">
            <div class="rect1"></div>
            <div class="rect2"></div>
            <div class="rect3"></div>
            <div class="rect4"></div>
            <div class="rect5"></div>
          </div>
        </div>
      </div>     
    </template> 
  </div>
  </loading>
</template>

<script>
import Loading from '@/components/Loading'
import dayjs from 'dayjs'
import utils from '@/utils'

export default {
  name: "Packet",
  components: {
    Loading
  },
  data () {
    return {
      loading: false,
      pktData: null,
      isClose: false,
      asset: {symbol: 'BTC'},
      lottery: null,
      user: null,
      greeting: '',
      openedCount: 0,
      totalCount: 0,
      openedAmount: 0,
      Amount: 0,
    }
  },
  computed: {
    hasAvatar () {
      return this.user && this.user.avatar_url
    },
    firstLetter () {
      return this.user ? this.user.full_name.trim()[0] : 'A'
    }
  },
  async mounted () {
    this.loading = true
    let pktId = this.$route.params.id
    let pktInfo = await this.GLOBAL.api.packet.show(pktId)
    if (pktInfo.error) {
      this.loading = false
      return 
    }
    let pktData = pktInfo.data
    for (var i in pktData.participants) {
      var p = pktData.participants[i];
      if (p.user_id === this.GLOBAL.api.account.userId()) {
        pktData.lottery = p;
        break;
      }
    }

    if (pktData.lottery || pktData.state === 'EXPIRED' || pktData.state === 'REFUNDED') {
      this.isClose = false
      for (var i in pktData.participants) {
        var participant = pktData.participants[i];
        pktData.participants[i]['symbol'] = pktData.asset.symbol;
        pktData.participants[i]['created_at'] = dayjs(pktData.participants[i].created_at).format('MM-DD HH:mm:ss')
        pktData.participants[i]['firstLetter'] = pktData.participants[i].avatar_url === '' ? (participant.full_name.trim()[0] || '^_^') : undefined;
      }
    } else {
      this.isClose = true
    }
    this.pktData = pktData
    this.user = pktData.user
    this.asset = pktData.asset
    this.lottery = pktData.lottery
    this.loading = false
  },
  methods: {
    async openPacket() {
      this.loading = true
      let packetId = this.$route.params.id
      let claimInfo = await this.GLOBAL.api.packet.claim(packetId)
      this.loading = false
      utils.reloadPage()
    }
  }
}
</script>

<style lang="scss" scoped>
@import '../assets/scss/constant.scss';
.packet-page {
  height: 100%;
  background: $color-main-highlight;
}
.packet-page.open {
  background: white;
}
.open {
  background: $color-main-highlight;
  .packet.header {
    padding-top: 64px;
  }
}
.open.button  {
  text-align: center;
  padding: 32px 0 64px;

  button {
    display: inline-block;
    background: $color-main-background;
    color: $color-main-highlight;
    font-size: 20px;
    font-weight: 300;
    border: 0 none;
    outline: 0 none;
    box-shadow: none;
    line-height: 1em;
    padding: 16px 32px;
    border-radius: 32px;
    cursor: pointer;
  }
}

.packet.header {
  box-sizing: border-box;
  background: $color-main-highlight;
  color: $color-main-background;
  font-family: $font-main-title;
  padding: 32px;
  text-align: center;
  overflow: hidden;

  .user.avatar {
    display: inline-block;
    width: 76px;
    height: 76px;
    border-radius: 38px;
    background-color: $color-main-background;
    color: $color-main-highlight;
    text-align: center;
    overflow: hidden;

    p {
      line-height: 76px;
      font-size: 32px;
      font-weight: 400;
      padding: 0;
      margin: 0;
    }

    img {
      width: 100%;
      height: 100%;
    }
  }

  .user.name {
    font-size: 24px;
    font-weight: 300;
    line-height: 1em;
    margin: 16px 0;
  }

  .greeting {
    font-size: 16px;
    font-weight: 100;
    line-height: 1em;
    margin: 0;
  }
}

.packet.lottery {
  box-sizing: border-box;
  color: $color-main-highlight;
  background: #EFEFEF;
  padding: 32px 16px;
  text-align: center;
  overflow: hidden;

  h3 {
    margin: 0;
    font-size: 36px;
    line-height: 1em;
    font-weight: 700;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;

    span {
      font-size: 16px;
      font-weight: 100;
      color: $color-main-foreground-dark;
    }
  }

  .expire.statement {
    font-size: 16px;
    font-weight: 500;
    line-height: 1.5em;
  }

  p {
    margin: 8px 0 0;
    font-size: 12px;
    line-height: 1em;
    color: $color-main-foreground-dark;
    opacity: 0.3;
  }
}

.packet.history {
  padding: 16px;

  h4 {
    margin: 0;
    font-size: 14px;
    line-height: 1em;
    font-weight: 100;
    padding: 0 0 16px;
  }

  ul {
    list-style: none;
    margin: 0;
    padding: 0;
    border-bottom: 1px solid #EFEFEF;
  }

  li {
    list-style: none;
    margin: 0;
    padding: 16px 0;
    border-top: 1px solid #EFEFEF;
    display: flex;
    width: 100%;

    .user.avatar {
      display: inline-block;
      min-width: 48px;
      width: 48px;
      height: 48px;
      border-radius: 24px;
      background: #EFEFEF;
      color: $color-main-highlight;
      text-align: center;
      overflow: hidden;

      p {
        line-height: 48px;
        font-size: 24px;
        font-weight: 400;
        padding: 0;
        margin: 0;
      }

      img {
        width: 100%;
        height: 100%;
      }
    }

    .user.info {
      box-sizing: border-box;
      padding: 0 16px;
      min-width: 0;
      font-family: $font-main-title;

      h5 {
        font-size: 20px;
        font-weight: 300;
        line-height: 1em;
        margin: 2px 0 8px;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }

      time {
        font-size: 12px;
        font-weight: 100;
        line-height: 1em;
        margin: 8px 0 0;
      }
    }

    .packet.amount {
      margin-left: auto;
      font-family: $font-main-mono;
      font-size: 20px;
      font-weight: 300;
      color: $color-main-highlight;
      white-space: nowrap;
    }
  }
}

</style>
