<template>
  <van-cell
    clickable
    :title="$t('withdraw.amount')"
    :class="{'is-required': required}"
    @click="show = !show"
  >
    <span v-if="value">{{ value }}</span>
    <span class="placeholder" v-else>{{ placeholder }}</span>
    <van-number-keyboard
      :show="show"
      :safe-area-inset-bottom="true"
      theme="custom"
      extra-key="."
      :close-button-text="$t('comm.complete')"
      @blur="show = false"
      @input="onInput"
      @delete="onDelete"
    />
  </van-cell>
</template>

<script>
export default {
  model: {
    prop: 'value',
    event: 'change'
  },
  props: ['value', 'placeholder', 'min', 'max', 'percision', 'required'],
  data() {
    return {
      show: false,
      numberArr: []
    }
  },
  computed: {
    numberPipe: {
      get() {
        return this.numberArr.join('')
      },
      set({ type, value = null }) {
        if (type === 'add') {
          if (this.checkInput(value)) {
            this.numberArr.push(value)
          }
        } else if (type === 'delete') {
          this.numberArr.pop(value)
        } else {
          throw 'unknown action'
        }
        this.$emit('change', this.numberPipe)
      }
    }
  },
  methods: {
    checkInput(value) {
      if (value ===  '.' && this.numberArr.indexOf('.') > -1) {
        return false
      }
      const tempValue = Number(this.numberPipe + value)
      if (isNaN(tempValue)) {
        return false
      }
      if (tempValue > this.max) {
        this.$toast(this.$t('errors.more_than_available'))
        return false
      }
      const dotIndex = this.numberArr.indexOf('.')
      const len = this.numberArr.length
      if (dotIndex > 0 && this.percision && (len - dotIndex) > this.percision) {
        return false
      }
      return true
    },
    onInput(value) {
      this.numberPipe = { type: 'add', value }
    },
    onDelete() {
      this.numberPipe = { type: 'delete' }
    }
  }
}
</script>

<style lang="scss" scoped>
.placeholder {
  color: #969799;
}
</style>


