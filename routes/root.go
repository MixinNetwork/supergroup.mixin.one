package routes

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/views"
	"github.com/dimfeld/httptreemux"
	"github.com/silenceper/wechat"
)

var wxcfg *wechat.Config
var wxclient *wechat.Wechat

func RegisterRoutes(router *httptreemux.TreeMux) {
	//配置微信参数
	wxcfg = &wechat.Config{
		AppID:          config.Get().Wechat.AppId,
		AppSecret:      config.Get().Wechat.AppSecret,
		Token:          config.Get().Wechat.Token,
		EncodingAESKey: config.Get().Wechat.EncodingAESKey,
	}
	wxclient = wechat.NewWechat(wxcfg)

	router.GET("/", root)
	router.GET("/_hc", healthCheck)
	router.GET("/wechat/request/:id", wxOAuthRequest)
	router.GET("/wechat/callback", wxOAuthCallback)
	registerUsers(router)
	registerPackets(router)
	registerMesseages(router)
	registerProperties(router)
	registerCoupons(router)
}

func root(w http.ResponseWriter, r *http.Request, params map[string]string) {
	views.RenderDataResponse(w, r, map[string]string{
		"build":      config.BuildVersion + "-" + runtime.Version(),
		"developers": "https://developers.mixin.one",
	})
}

func healthCheck(w http.ResponseWriter, r *http.Request, params map[string]string) {
	views.RenderBlankResponse(w, r)
}

func wxOAuthRequest(w http.ResponseWriter, r *http.Request, params map[string]string) {
	userId := params["id"]
	wxoauth := wxclient.GetOauth()
	url, err := wxoauth.GetRedirectURL(config.Get().Service.HTTPResourceHost+"/wechat/callback", "snsapi_userinfo", userId)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(url)
	http.Redirect(w, r, url, 302)
}

func wxOAuthCallback(w http.ResponseWriter, r *http.Request, params map[string]string) {
	//配置微信参数
	wxoauth := wxclient.GetOauth()
	code := r.URL.Query().Get("code")
	userId := r.URL.Query().Get("state")
	resToken, err := wxoauth.GetUserAccessToken(code)
	if err != nil {
		fmt.Println(err)
		return
	}
	url := fmt.Sprintf(config.Get().Service.HTTPResourceHost+"/?#/wxpay?access_token=%s&open_id=%s&user_id=%s", resToken.AccessToken, resToken.OpenID, userId)
	http.Redirect(w, r, url, 302)
}
