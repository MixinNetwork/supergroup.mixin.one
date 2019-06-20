<template>
  <div class="members-page">
    <nav-bar title="Members" :hasTopRight="false" :hasBack="true"></nav-bar>
    <van-panel title="开发中">
    </van-panel>
  </div>
</template>

<script>
import NavBar from '@/components/Nav'
export default {
  name: 'Members',
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
.members-page {
  padding-top: 60px;
}
</style>
