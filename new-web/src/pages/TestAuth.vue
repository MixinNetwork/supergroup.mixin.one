<template>
  <div></div>
</template>

<script>
import { Toast } from 'vant'

export default {
  async mounted() {
    const code = this.$route.query.code
    Toast('Loading')
    try {
      let resp = await this.GLOBAL.api.account.authenticate(code)
      if (resp.data.authentication_token) {
        this.$router.push('/')
      }
    } catch (err) {
      console.log(err)
      Toast('Failed to Authorize')
      this.$router.push('/')
    }
  }
}
</script>
