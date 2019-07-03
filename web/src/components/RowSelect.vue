<template>
  <div class="row-select">
    <van-cell :title="title" @click="show = true" is-link>
      <slot name="text">
      </slot>
    </van-cell>
    <div></div>
    <van-popup
        v-model="show"
        position="bottom"
      >
        <van-picker
          show-toolbar
          :columns="columns"
          @confirm="handleConfirm"
          @cancel="show = false"
        />
      </van-popup>
  </div>
</template>

<script>
export default {
  model: {
    prop: 'index',
    event: 'change'
  },
  props: ['index', 'title', 'columns', 'placeholder'],
  data() {
    return {
      show: false,
    }
  },
  methods: {
    handleConfirm(value, index) {
      this.$emit('change', index)
      this.show = false
    }
  }
}
</script>
