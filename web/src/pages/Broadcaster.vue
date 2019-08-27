<template>
  <loading :loading="loading" :fullscreen="true">
  <div class="broadcaster-page">
    <nav-bar :title="$t('broadcaster.title')" :hasTopRight="false" :hasBack="true"></nav-bar>
    <van-cell v-if="role=='admin'">
      <van-field placeholder="Add Broadcaster By Identity Number"
        @change="addBroadcaster" v-model="broadcasterInput"
        >
      </van-field>
    </van-cell>
    <van-list class="broadcasters">
    <div :key="item.user_id" v-for="item in broadcasters" class="broadcaster" @click="selectBroadcaster(item)">
      <img class="member-icon-img" :src="item.avatar_url"/>
      <div class="full-name">
        {{item.full_name}}
      </div>
      <span v-if="broadcaster.user_id==item.user_id" class="current">&#10004;</span>
    </div>
    </van-list>

    <van-cell-group class='reward' title="">
      <row-select
        :index="0"
        :title="$t('prepare_packet.select_assets')"
        :columns="assets"
        placeholder="Tap to Select"
        @change="onChangeAsset">
        <span slot="text">{{selectedAsset ? selectedAsset.text : 'Tap to Select'}}</span>
      </row-select>
      <van-cell>
        <van-field type="number" v-model="amount" :label="$t('prepare_packet.amount')" :placeholder="$t('prepare_packet.placeholder_amount')">
          <span slot="right-icon">{{selectedAsset ? selectedAsset.symbol : ''}}</span>
        </van-field>
      </van-cell>
    </van-cell-group>
    <van-row style="padding: 20px">
      <van-col span="24">
        <van-button style="width: 100%" type="info" :disabled="!validated" @click="pay">{{$t('broadcaster.pay')}}</van-button>
      </van-col>
    </van-row>
  </div>
  </loading>
</template>

<script>
import Loading from '@/components/Loading'
import NavBar from '@/components/Nav'
import RowSelect from '@/components/RowSelect'
import uuid from 'uuid'
import { CLIENT_ID } from '@/constants'

export default {
  name: 'Broadcaster',
  data () {
    return {
      loading: false,
      broadcasterInput: '',
      broadcasters: [],
      broadcaster: {},
      assets: [],
      selectedAsset: null,
      amount: '',
      role: 'memeber',
    }
  },
  components: {
    NavBar, RowSelect, Loading
  },
  async mounted () {
    this.loading = true;
    this.role = window.localStorage.getItem('role');
    let broadcasters = await this.GLOBAL.api.broadcaster.index();
    if (broadcasters.data) {
      this.broadcasters = broadcasters.data;
      if (this.broadcasters.length > 0) {
        this.broadcaster = this.broadcasters[0];
      }
    }
    let assets = await this.GLOBAL.api.broadcaster.assets();
    if (assets.data) {
      this.assets = assets.data.map((x) => {
        x.text = `${x.symbol} (${x.balance})`;
        return x;
      });
    }
    this.loading = false
  },
  computed: {
    validated () {
      if (this.broadcaster.user_id && this.amount && this.selectedAsset) {
        return true
      }
      return false
    }
  },
  methods: {
    selectBroadcaster (member) {
      this.broadcaster = member;
    },
    onChangeAsset (ix) {
      this.selectedAsset = this.assets[ix]
    },
    addBroadcaster () {
      this.GLOBAL.api.broadcaster.create(this.broadcasterInput).then((resp) => {
        if (resp.data) {
          this.broadcasterInput = '';
          for (let i=0;i<this.broadcasters.length;i++) {
            if (this.broadcasters[i].user_id == resp.data.user_id) {
              return
            }
          }
          this.broadcasters.unshift(resp.data);
        }
      });
    },
    async pay () {
      let memo = btoa(`REWARD:${this.broadcaster.user_id}`);
      let traceId = uuid.v4();
      let amount = this.amount;
      window.location.href = `mixin://pay?recipient=${CLIENT_ID}&asset=${this.selectedAsset.asset_id}&trace=${traceId}&amount=${amount}&memo=${memo}`;
      this.amount = '';
    }
  }
}
</script>

<style scoped>
.broadcaster {
  display: flex;
  align-items: center;
  border-bottom: 1px solid #f8f8f8;
  padding: 6px 10px;
}
.broadcaster-page {
  padding-top: 60px;
}
.broadcasters {
  background: white;
  margin-top: 16px;
  padding: 16px 6px;
}
.current {
color: #1989fa;
}
.member-icon-img {
  width: 28px;
  height: 28px;
  border-radius: 14px;
}
.full-name {
  padding-left: 6px;
  flex-grow: 1;
}
.reward {
  margin-top: 16px;
}
</style>
