package routes

import (
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
		AppID:          config.AppConfig.Wechat.AppId,
		AppSecret:      config.AppConfig.Wechat.AppSecret,
		Token:          config.AppConfig.Wechat.Token,
		EncodingAESKey: config.AppConfig.Wechat.EncodingAESKey,
	}
	wxclient = wechat.NewWechat(wxcfg)

	router.GET("/", root)
	router.GET("/_hc", healthCheck)
	registerUsers(router)
	registerPackets(router)
	registerMesseages(router)
	registerProperties(router)
	registerCoupons(router)
	registerWechat(router)
	registerPluginRoutes(router)
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
