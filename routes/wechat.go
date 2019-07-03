package routes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/models"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/MixinNetwork/supergroup.mixin.one/views"
	"github.com/dimfeld/httptreemux"
	"github.com/objcoding/wxpay"
)

type wechatImpl struct{}

func registerWechat(router *httptreemux.TreeMux) {
	impl := &wechatImpl{}
	router.POST("/wechat/pay/create", impl.createWxPay)
	router.POST("/wechat/pay/callback", impl.wxPayCallback)
	router.GET("/wechat/pay/:id", impl.checkWxPay)
	router.GET("/wechat/request/:id", impl.wxOAuthRequest)
	router.GET("/wechat/callback", impl.wxOAuthCallback)
}

func (impl *wechatImpl) wxPayCallback(w http.ResponseWriter, r *http.Request, params map[string]string) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	fmt.Printf("wxPayCallback: %v\n", bodyString)
	views.RenderDataResponse(w, r, config.GetExported())
}

func (impl *wechatImpl) createWxPay(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	var payload struct {
		OpenId string `json:"open_id"`
		UserId string `json:"user_id"`
	}
	var resp struct {
		Order       *models.Order `json:"order"`
		PayParams   wxpay.Params  `json:"pay_params"`
		PayJsParams wxpay.Params  `json:"pay_js_params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
		return
	}
	if order, payParams, payJsParams, err := models.CreateOrder(r.Context(), payload.UserId, "19.9", payload.OpenId); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		resp.Order = order
		resp.PayParams = payParams
		resp.PayJsParams = payJsParams
		views.RenderDataResponse(w, r, resp)
	}
}

func (impl *wechatImpl) checkWxPay(w http.ResponseWriter, r *http.Request, params map[string]string) {
	id := params["id"]
	if s, err := models.GetOrder(r.Context(), id); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, s)
	}
}

func (impl *wechatImpl) wxOAuthRequest(w http.ResponseWriter, r *http.Request, params map[string]string) {
	userId := params["id"]
	wxoauth := wxclient.GetOauth()
	url, err := wxoauth.GetRedirectURL(config.AppConfig.Service.HTTPResourceHost+"/wechat/callback", "snsapi_userinfo", userId)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(url)
	http.Redirect(w, r, url, 302)
}

func (impl *wechatImpl) wxOAuthCallback(w http.ResponseWriter, r *http.Request, params map[string]string) {
	//配置微信参数
	wxoauth := wxclient.GetOauth()
	code := r.URL.Query().Get("code")
	userId := r.URL.Query().Get("state")
	resToken, err := wxoauth.GetUserAccessToken(code)
	if err != nil {
		fmt.Println(err)
		return
	}
	url := fmt.Sprintf(config.AppConfig.Service.HTTPResourceHost+"/?#/wxpay?access_token=%s&open_id=%s&user_id=%s", resToken.AccessToken, resToken.OpenID, userId)
	http.Redirect(w, r, url, 302)
}
