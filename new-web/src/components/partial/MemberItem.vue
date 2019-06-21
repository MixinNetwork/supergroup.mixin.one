<template>
  <div class="member-item" @click="handleMemberClick(member)">
    <div class="cell member-icon">
      <img class="member-icon-img" :src="member.avatar_url"/>
    </div>
    <div class="cell member-list-info">
      <div class="member-name">{{ member.full_name }}</div>
        <div class="member-id">{{ member.identity_number }}</div>
    </div>
    <div class="cell member-list-role">
      <div class="member-role" :class="member.role === 'admin' ? member.role : ''"></div>
      <div class="member-time">{{ member.time }}</div>
    </div>
  </div>
</template>

<script>
export default {
  props: ['member'],
  methods: {
    handleMemberClick(member) {
      this.$emit('member-click', member)
    }
  }
}
</script>

<style lang="scss" scoped>
.member-item {
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

  .member-icon {
    padding-right: 10px;

    .member-icon-img {
      width: 28px;
      height: 28px;
      border-radius: 14px;
    }
  }

  .member-list-info {
    flex: 1;
  }
  .member-list-role {
    align-items: flex-end;
  }

  .member-name, .member-role {
    height: 20px;
    font-size: 14px;
  }

  .member-role.admin {
    width: 22px;
    background: url('../../assets/images/icon-admin.png') 0 0 no-repeat;
    background-size: 22px 21px;
  }

  .member-id, .member-time {
    opacity: 0.6;
  }
}
</style>

