<template>
  <div class="message-item" @click="handleClick(message)">
    <div class="cell message-list-info">
      <div class="message-name">{{ message.full_name }}</div>
        <div class="message-id">{{ message.data || ' No Data ' }}</div>
    </div>
    <div class="cell message-list-role">
      <div class="message-role">{{message.category}}</div>
      <div class="message-time">{{ message.time }}</div>
    </div>
  </div>
</template>

<script>
export default {
  props: ['message'],
  methods: {
    handleClick(message) {
      this.$emit('message-click', message);
    }
  }
}
</script>

<style lang="scss" scoped>
.message-item {
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

  .message-list-info {
    flex: 1;
  }
  .message-list-role {
    align-items: flex-end;
  }

  .message-name, .message-role {
    height: 20px;
    font-size: 14px;
  }

  .message-role.admin {
    width: 22px;
    background: url('../../assets/images/icon-admin.png') 0 0 no-repeat;
    background-size: 22px 21px;
  }

  .message-id, .message-time {
    opacity: 0.6;
  }
}
</style>

