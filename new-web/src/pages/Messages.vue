<template>
  <div class="messages-page">
    <nav-bar title="Messages" :hasTopRight="false" :hasBack="true"></nav-bar>
    <van-panel title="开发中">
    </van-panel>
  </div>
</template>

<script>
import NavBar from '@/components/Nav'
export default {
  name: 'Messages',
  props: {
  },
  data () {
    return {
    }
  },
  components: {
    NavBar
  },
  methods: {
  }
}
</script>

<style scoped>
.messages-page {
  padding-top: 60px;
}
</style>
