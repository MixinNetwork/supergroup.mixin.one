<template>
  <div id="app">
    <router-view></router-view>
  </div>
</template>

<script>
export default {
  name: 'App',
  components: {
  },
  data() {
    return {
    }
  },
  mounted() {
    this.GLOBAL.api.net.on(401, (payload)=>{
      const CLIENT_ID = window.localStorage.getItem('cfg_client_id')
      const OAUTH_CALLBACK_URL = window.localStorage.getItem('cfg_host' + '/#/auth')
      let url = `https://mixin.one/oauth/authorize?client_id=${CLIENT_ID}&scope=PROFILE:READ+ASSETS:READ&response_type=code&return_to=${encodeURIComponent(OAUTH_CALLBACK_URL)}` 
      window.location.replace(url)
    })
  },
  methods: {
  }
}
</script>

<style>
html, body {
  padding: 0;
  margin: 0;
  height: 100%;
  font-size: 14px;
}
a {
}

#app {
  background: #f8f8f8;
  font-family: 'Roboto', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
  padding: 0;
  height: 100%;
}
.van-cell.van-field {
  padding: 0;
}
</style>
