<template>
  <div></div>
</template>

<script>
import { Toast } from 'vant'

export default {
  async mounted() {
    const code = this.$route.query.code
    try {
      let resp = await this.GLOBAL.api.account.authenticate(code)
      if (resp.data.authentication_token) {
        this.$router.push('/')
      }
    } catch (err) {
      Toast('OAuth Failed')
      this.$router.push('/')
    }
  }
}
</script>
