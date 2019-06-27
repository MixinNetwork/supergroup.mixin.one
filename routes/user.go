package routes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/middlewares"
	"github.com/MixinNetwork/supergroup.mixin.one/models"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/MixinNetwork/supergroup.mixin.one/views"
	"github.com/dimfeld/httptreemux"
	"github.com/objcoding/wxpay"
)

type usersImpl struct{}

type userRequest struct {
	FullName string `json:"full_name"`
}

func registerUsers(router *httptreemux.TreeMux) {
	impl := &usersImpl{}
	router.POST("/auth", impl.authenticate)
	router.POST("/account", impl.update)
	router.POST("/subscribe", impl.subscribe)
	router.POST("/unsubscribe", impl.unsubscribe)
	router.POST("/users/:id/remove", impl.remove)
	router.POST("/users/:id/block", impl.block)
	router.POST("/wechat/pay/create", impl.createWxPay)
	router.POST("/wechat/pay/callback", impl.wxPayCallback)
	router.GET("/wechat/pay/:id", impl.checkWxPay)
	router.GET("/me", impl.me)
	router.GET("/subscribers", impl.subscribers)
	router.GET("/users/:id", impl.show)
	router.GET("/amount", impl.amount)
	router.GET("/config", impl.getConfig)
}

func (impl *usersImpl) authenticate(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if user, err := models.AuthenticateUserByOAuth(r.Context(), body.Code); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderAccount(w, r, user)
	}
}

func (impl *usersImpl) update(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body userRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
		return
	}
	current := middlewares.CurrentUser(r)
	if err := current.UpdateProfile(r.Context(), body.FullName); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderAccount(w, r, current)
	}
}

func (impl *usersImpl) me(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	views.RenderAccount(w, r, middlewares.CurrentUser(r))
}

func (impl *usersImpl) subscribers(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	offset, _ := time.Parse(time.RFC3339Nano, r.URL.Query().Get("offset"))
	var num int64
	var keywords string
	var err error
	num, err = strconv.ParseInt(r.URL.Query().Get("q"), 10, 64)
	if err != nil {
		keywords = r.URL.Query().Get("q")
	}
	if users, err := models.Subscribers(r.Context(), offset, num, keywords); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderUsersView(w, r, users)
	}
}

func (impl *usersImpl) subscribe(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	current := middlewares.CurrentUser(r)
	if err := current.Subscribe(r.Context()); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderAccount(w, r, current)
	}
}

func (impl *usersImpl) unsubscribe(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	current := middlewares.CurrentUser(r)
	if err := current.Unsubscribe(r.Context()); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderAccount(w, r, current)
	}
}

func (impl *usersImpl) remove(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if err := middlewares.CurrentUser(r).DeleteUser(r.Context(), params["id"]); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderBlankResponse(w, r)
	}
}

func (impl *usersImpl) block(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if _, err := middlewares.CurrentUser(r).CreateBlacklist(r.Context(), params["id"]); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderBlankResponse(w, r)
	}
}

func (impl *usersImpl) show(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if user, err := models.FindUser(r.Context(), params["id"]); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else if user == nil {
		views.RenderErrorResponse(w, r, session.NotFoundError(r.Context()))
	} else {
		views.RenderUserView(w, r, user)
	}
}

func (impl *usersImpl) amount(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	if s, err := models.ReadStatistic(r.Context(), middlewares.CurrentUser(r)); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, s)
	}
}

func (impl *usersImpl) getConfig(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	views.RenderDataResponse(w, r, config.GetExported())
}

func (impl *usersImpl) wxPayCallback(w http.ResponseWriter, r *http.Request, params map[string]string) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	fmt.Printf("wxPayCallback: %v\n", bodyString)
	views.RenderDataResponse(w, r, config.GetExported())
}

func (impl *usersImpl) createWxPay(w http.ResponseWriter, r *http.Request, _ map[string]string) {
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

func (impl *usersImpl) checkWxPay(w http.ResponseWriter, r *http.Request, params map[string]string) {
	id := params["id"]
	if s, err := models.GetOrder(r.Context(), id); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, s)
	}
}
