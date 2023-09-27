<template>
  <loading :loading="false" :fullscreen="true">
  <div class="container">
    <nav-bar :title="$t('proof.title')"></nav-bar>
    <van-panel>
      <div>
        <van-cell>
          {{$t('proof.desc')}}
          <van-button style="width: 100%" type="primary" @click="joinAgain">{{$t('proof.join')}}</van-button>
        </van-cell>
      </div>
    </van-panel>
  </div>
  </loading>
</template>

<script>
import {CLIENT_ID, MIXIN_AUTH_API} from '@/constants'
import NavBar from '@/components/NavBar'
import Loading from '../components/LoadingSpinner'

export default {
  name: 'PayPage',
  components: {
    NavBar, Loading
  },
  methods: {
    joinAgain () {
      let url = `${MIXIN_AUTH_API}/oauth/authorize?client_id=${CLIENT_ID}&scope=PROFILE:READ+ASSETS:READ+MESSAGES:REPRESENT&response_type=code`
      window.location.href = url ;
    },
  },
}
</script>

<style scoped>
.container {
  padding-top: 60px;
}
</style>
