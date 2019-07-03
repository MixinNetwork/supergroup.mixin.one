<template>
  <!-- <el-row v-if="hasBack || hasTopRight" class="nav">
    <el-col :span="3">
      <el-button v-show="hasBack" type="text" icon="el-icon-back" class="back-button" @click="back"></el-button>
    </el-col>
    <el-col :span="hasBack ? 18: 24"><span class="title">{{title}}</span></el-col>
    <el-col :span="4"></el-col>
  </el-row> -->
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
  name: 'Nav',
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
    return {
    }
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
  // padding: 0px 20px;
  left: 0;
  right: 0;
  top: 0;
  z-index: 9;
  background: rgba(255, 255, 255, 0.9);
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
  // height: 44px;
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
