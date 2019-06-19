<template>
  <div class="snapshot-item" @click="handleSnapshotClick(snapshot)">
    <div class="cell transaction-list-info">
      <div class="transaction-opponent">{{ snapshot.type || '- -' }}</div>
      <div class="transaction-time">{{ format(snapshot.created_at)  }}</div>
    </div>
    <div class="cell transaction-list-amount">
      <div class="transaction-amount">{{ snapshot.amount }}</div>
      <div class="transaction-amount-usd">${{ snapshot.amountUsd }}</div>
    </div>
  </div>
</template>

<script>
import { format } from '@/utils/dateFormat'
export default {
  props: ['snapshot'],
  methods: {
    handleSnapshotClick(snapshot) {
      this.$emit('snapshot-click', snapshot)
    },
    format
  }
}
</script>

<style lang="scss" scoped>
.snapshot-item {
  display: flex;
  align-items: center;

  .cell {
    display: flex;
    flex-direction: column;
  }

  .transaction-list-info {
    flex: 1;
  }

  .transaction-list-amount {
    align-items: flex-end;
  }

  .transaction-opponent, .transaction-amount {
    font-size: 18px;
  }
  
  .transaction-time, .transaction-amount-usd {
    opacity: 0.6;
  }
}
</style>
