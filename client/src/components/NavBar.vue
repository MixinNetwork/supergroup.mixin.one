<template>
  <van-nav-bar
    class="nav"
    :title="title"
    :left-text="hasBack ? $t('comm.back') : ''"
    :right-text="hasTopRight ? (this.rightText || $t('comm.more')) : ''"
    :left-arrow="hasBack"
    @click-left="back"
    @click-right="more"
  />
</template>

<script>
import { NavBar } from 'vant';
export default {
  name: 'NavBar',
  props: {
    title: String,
    rightText: String,
    hasBack: Boolean,
    hasTopRight: {
      type: Boolean,
      default: false
    },
  },
  components: {
    'van-nav-bar': NavBar,
  },
  data() {
    return {}
  },
  methods: {
    back () {
      this.$router.back()
    },
    more () {
      this.$emit("click-right", {})
    }
  }
}
</script>

<style lang="scss" scoped>
.nav {
  position: fixed;
  left: 0;
  right: 0;
  top: 0;
  z-index: 9;
  background: rgba(255, 255, 255, 0.9);
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
  .title {
    font-size: 16px;
    height: 44px;
    display: flex;
    justify-content: center;
    align-items: center;
  }
  .back-button {
    font-size: 20px;
    height: 44px;
    height: 44px;
    padding: 0 20px;
    margin-left: -20px;
    font-weight: bold;
  }
}
</style>
