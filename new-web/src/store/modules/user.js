import Vue from 'vue'
import UserService from '@/service/user'
import MixinService from '@/service/mixin'
import { GETTERS, ACTIONS, MUTATIONS } from '../keys'
import toPrecision from '@/utils/toPrecision'

const state = {
  profile: null,
  assets: [],
  assetsIdMap: {},
  addrs: [],
  addrsIdMap: {},

}

const getters = {
  [GETTERS.USER.TOTAL_BTC]: ({ assets, assetsIdMap }) => id => {
    if (id) {
      const { price_btc, balance } = assetsIdMap[id]
      return toPrecision(price_btc * balance, 8)
    }
    const total = assets.reduce((acc, asset) => {
      const { price_btc, balance } = asset
      return price_btc * balance + acc
    }, 0)
    return toPrecision(total, 8)
  },

  [GETTERS.USER.TOTAL_USD]: ({ assets, assetsIdMap }) => id => {
    if (id) {
      const { price_usd, balance } = assetsIdMap[id]
      return toPrecision(price_usd * balance, 2)
    }
    const total = assets.reduce((acc, asset) => {
      const { price_usd, balance } = asset
      return price_usd * balance + acc
    }, 0)
    return toPrecision(total, 2)
  },

  [GETTERS.USER.ASSET]: ({ assetsIdMap }) => id => {
    const asset = assetsIdMap[id]
    const { price_usd, balance } = asset
    const amountUsd = toPrecision(price_usd * balance, 2)
    return {
      ...asset,
      amountUsd
    }
  },

  [GETTERS.USER.ADDRESSES]: ({ addrs }) => () => {
    // const addr = addrsIdMap[id]
    return addrs
  }

}

const mutations = {
  [MUTATIONS.USER.SET_PROFILE]: (state, profile) => {
    state.profile = profile
  },

  [MUTATIONS.USER.SET_ASSETS]: (state, assets) => {
    const map = {}
    state.assets = assets.map(asset => {
      const { price_usd, balance, asset_id } = asset
      map[asset_id] = asset
      return { ...asset, amountUsd: toPrecision(price_usd * balance, 2) }
    }).sort((a, b) => {
      return b.amountUsd - a.amountUsd
    })
    Vue.set(state, 'assetsIdMap', map)
  },

  [MUTATIONS.USER.SET_ADDRS]: (state, addrs) => {
    const map = {}
    state.addrs = addrs.map(addr => {
      const { address_id } = addr
      map[address_id] = addr
      return addr
    })
    Vue.set(state, 'addrsIdMap', map)
  }

}

const actions = {

  [ACTIONS.USER.GET_PROFILE]: async ({ commit }) => {
    const res = await UserService.getProfile()
    commit(MUTATIONS.USER.SET_PROFILE, res)
  },

  [ACTIONS.USER.GET_ASSETS]: async ({ commit }) => {
    const res = await MixinService.getAssets()
    commit(MUTATIONS.USER.SET_ASSETS, res)
  },

  [ACTIONS.USER.GET_ADDRS]: async ({ commit }, assetId) => {
    const res = await MixinService.getAssetAddresses(assetId)
    commit(MUTATIONS.USER.SET_ADDRS, res)
  }

}

export default {
  state,
  mutations,
  getters,
  actions
}
