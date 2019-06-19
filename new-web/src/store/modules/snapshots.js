import { GETTERS, ACTIONS, MUTATIONS } from '../keys'
import MixinService from '@/service/mixin'
import toPrecision from '@/utils/toPrecision'

const state = {
  snapshots: [],
}

const getters = {
  [GETTERS.SNAPSHOTS.SNAPSHOTS]: (state, _, rootState) => {
    const { user: { assetsIdMap } } = rootState
    const { snapshots } = state
    const transfered = snapshots.map(snapshot => {
      try {
        const { asset_id, amount } = snapshot
        const { price_usd } = assetsIdMap[asset_id]
        const amountUsd = toPrecision(amount * price_usd, 2)
        return {
          ...snapshot,
          amountUsd
        }
      } catch (e) {
        return snapshot
      }
    })
    return transfered
  }
}

const mutations = {
  [MUTATIONS.SNAPSHOTS.SET_SNAPSHOTS]: (state, snapshots) => {
    state.snapshots = state.snapshots.concat(snapshots)
  },
  [MUTATIONS.SNAPSHOTS.REMOVE_SNAPSHOTS]: (state) => {
    state.snapshots = []
  },
}

const actions = {
  [ACTIONS.SNAPSHOTS.GET_SNAPSHOTS]: async ({ commit, state }, { id }) => {
    const { snapshots: oldSnapshots } = state
    const lasted = oldSnapshots[oldSnapshots.length - 1]
    const offset = lasted && lasted['created_at'] 
    const params = { asset: id, limit: 10, offset }
    const snapshots = await MixinService.getSnapshots(params)
    commit(MUTATIONS.SNAPSHOTS.SET_SNAPSHOTS, snapshots )
    return snapshots
  }
}

export default {
  state,
  getters,
  mutations,
  actions
}