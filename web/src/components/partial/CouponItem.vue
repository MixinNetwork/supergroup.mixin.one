<template>
  <div class="coupon-item" @click="handleClick(coupon)">
    <div class="cell coupon-list-info">
      <div class="coupon-code">{{ coupon.code }}</div>
        <div class="coupon-id"></div>
    </div>
    <div class="cell coupon-list-role">
      <div class="coupon-user">{{ coupon.full_name ? $t('coupons.state_used', {name: coupon.full_name}) : $t('coupons.state_available')}}</div>
      <div class="coupon-time">{{ coupon.time }}</div>
    </div>
  </div>
</template>

<script>
export default {
  props: ['coupon'],
  methods: {
    handleClick(coupon) {
      this.$emit('coupon-click', coupon)
    }
  }
}
</script>

<style lang="scss" scoped>
.coupon-item {
  display: flex;
  font-size: 12px;
  padding: 6px 10px;
  border-bottom: 1px solid #f8f8f8;
  background: #fff;
  
  .cell {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    justify-content: center;
  }

  .coupon-list-info {
    flex: 1;
  }
  .coupon-list-role {
    align-items: flex-end;
  }

  .coupon-code {
    font-size: 16px;
  }
  .coupon-user {
    height: 20px;
    font-size: 14px;
  }

  .coupon-user.admin {
    width: 22px;
    background: url('../../assets/images/icon-admin.png') 0 0 no-repeat;
    background-size: 22px 21px;
  }

  .coupon-id, .coupon-time {
    opacity: 0.6;
  }
}
</style>

