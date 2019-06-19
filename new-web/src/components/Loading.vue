<template>
  <div class="mask" :class="{'full-screen': fullscreen, 'is-loading': loading}">
    <van-loading v-if="loading" type="spinner"  color="white"  class="loading"/>
    <slot />
  </div>
</template>

<script>
export default {
  props: ['loading', 'fullscreen']
}
</script>

<style lang="scss" scoped>
.mask {
  position: relative;
  width: 100%;
  height: 100%;
  min-height: 200px;

  &.is-loading {
    pointer-events: none;
  }

  &.full-screen {
    position: absolute;
    top: 0;
    left: 0;
  }

  .loading {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    z-index: 1;
    padding: 10px;
    background: rgba(0, 0, 0, 0.8);
    border-radius: 3px;
  }
}
</style>
