export const isProduction = process.env.NODE_ENV === 'production'

export const CLIENT_ID = process.env.VUE_APP_CLIENT_ID

export const MIXIN_AUTH_API = process.env.VUE_APP_MIXIN_AUTH_API

export const WEB_ROOT = process.env.VUE_APP_WEB_ROOT

export const OAUTH_CALLBACK_URL = process.env.VUE_APP_WEB_ROOT + '/auth'

export const BASE_URL = process.env.VUE_APP_API_ROOT

export const ROUTER_MODE = process.env.VUE_APP_ROUTER_MODE || 'history'

export const MIXIN_HOST = 'https://api.mixin.one'

export const EOS_ASSET_ID = '6cfe566e-4aad-470b-8c9a-2fd35b49c68d'
