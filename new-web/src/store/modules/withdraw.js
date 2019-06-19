import { MUTATIONS } from '../keys'

const state = {
  selectAssetIndex: null,
  selectAssetAddressIndex: null
}

const mutations = {
  [MUTATIONS.WITHDRAW.SET_ASSET_INDEX](state, index) {
    state.selectAssetIndex = index
  },
  [MUTATIONS.WITHDRAW.SET_ASSET_ADDR_INDEX](state, index) {
    state.selectAssetAddressIndex = index
  }
}

export default {
  state,
  mutations
}