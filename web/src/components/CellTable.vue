<template>
  <div class="cell-table">
    <van-row class="table-row" v-for="group in groupedItems">
      <van-col span="6" v-for="item in group">
        <a v-if="item.click" class="item" @click="item.click">
          <img :src="item.icon"/>
          <span>{{item.label}}</span>
        </a>
        <a v-else-if="item.url.indexOf('http') === 0 || item.isPlugin" class="item" @click="openExternalLink(item.url)">
          <img :src="item.icon"/>
          <span>{{item.label}}</span>
        </a>
        <router-link v-else="item.url" class="item" :to="item.url">
          <img :src="item.icon"/>
          <span>{{item.label}}</span>
        </router-link>
      </van-col>
    </van-row>
  </div>
</template>

<script>
export default {
  name: 'CellTable',
  props: {
    items: {
      type: Array,
      default: []
    },
  },
  data() {
    return {
    }
  },
  computed: {
    groupedItems () {
      let groups = []
      for (let ix = 0; ix < this.items.length; ix += 4) {
        groups.push(this.items.slice(ix, ix + 4))
      }
      return groups
    }
  },
  methods: {
    openExternalLink (url) {
      window.location.href = url
    }
  }
}
</script>

<style lang="scss" scoped>
.cell-table {
  padding: 20px 10px 2px 10px;
}
.table-row {
  margin-bottom: 16px;
}
.item {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  img {
    border-radius: 99em;
    max-height: 48px;
    max-width: 48px;
    width: 100%;
    height: 100%;
  }
  span {
    padding: 4px 0;
    color: #333;
    text-align: center;
    font-size: 12px;
  }
}
</style>
