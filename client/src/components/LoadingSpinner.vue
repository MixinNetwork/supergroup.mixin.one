<template>
  <div class="mask" :class="{'full-screen': fullscreen, 'is-loading': loading}">
    <van-loading v-if="loading" color="#1989fa" class="loading"/>
    <slot />
    <div v-if="loading" class="mask-shadow" :class="{'full-screen': fullscreen, 'is-loading': loading}"></div>
  </div>
</template>

<script>
export default {
  props: ['loading', 'fullscreen']
}
</script>

<style lang="scss" scoped>
.mask-shadow {
  width: 100%;
  height: 100%;
  background: rgba(255, 255, 255, 0.5);

  &.is-loading {
    pointer-events: none;
  }

  &.full-screen {
    position: absolute;
    top: 0;
    left: 0;
  }
}
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
    border-radius: 3px;
  }
}
</style>
