<template>
  <div class="row-select">
    <van-cell :title="title" @click="show = true" is-link>
      <slot>
      </slot>
    </van-cell>
    <div></div>
    <van-action-sheet v-model:show="show" :actions="columns" @select="onSelect">
    </van-action-sheet>
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
    onSelect(value, index) {
      this.$emit('change', index)
      this.show = false
    }
  },
}
</script>
