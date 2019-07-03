<template>
  <div class="asset-item" @click="handleAssetClick(asset)">
    <div class="cell asset-icon">
      <img class="asset-icon-img" :src="asset.icon_url"/>
    </div>
    <div class="cell asset-list-info">
      <div class="asset-symbol">{{ asset.symbol }}</div>
        <div class="asset-name">{{ asset.name }}</div>
    </div>
    <div class="cell asset-list-amount">
      <div class="asset-amount">{{ asset.balance }}</div>
      <div class="asset-amount-usd">${{ asset.amountUsd }}</div>
    </div>
  </div>
</template>

<script>
export default {
  props: ['asset'],
  methods: {
    handleAssetClick(asset) {
      this.$emit('asset-click', asset)
    }
  }
}
</script>

<style lang="scss" scoped>
.asset-item {
  display: flex;
  
  .cell {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    justify-content: center;
  }

  .asset-icon {
    padding-right: 10px;

    .asset-icon-img {
      width: 42px;
      height: 42px;
    }
  }

  .asset-list-amount {
    flex: 1;
    align-items: flex-end;
  }

  .asset-symbol, .asset-amount {
    font-size: 18px;
  }

  .asset-name, .asset-amount-usd {
    opacity: 0.6;
  }
}
</style>

