<template>
  <div class="address-item" @click="handleClick(address)">
    <div v-if="address.account_name && address.account_tag" class="cell">
      <div class="label">{{ address.account_name }}</div>
      <div class="publickey">{{ address.account_tag }}</div>
    </div>
    <div class="cell">
      <div class="label">{{ address.label }}</div>
      <div class="publickey">{{ address.public_key }}</div>
    </div>
  </div>
</template>

<script>
export default {
  props: ['address'],
  methods: {
    handleClick(address) {
      this.$emit('address-click', address)
    }
  }
}
</script>

<style lang="scss" scoped>
.address-item {
  display: flex;
  
  .cell {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    justify-content: center;
  }

  .label {
    font-size: 16px;
  }
  .publickey {
    font-size: 14px;
    opacity: 0.6;
    font-size: 14px;
    opacity: 0.8;
    word-break: break-all;
    line-height: 1.3;
  }
}
</style>

