package main

import (
	"fmt"
	"net/http"

	"github.com/MixinNetwork/supergroup.mixin.one/middlewares"
	"github.com/MixinNetwork/supergroup.mixin.one/models"
	"github.com/MixinNetwork/supergroup.mixin.one/plugin"
	"github.com/gin-gonic/gin"
)

var pluginContext *plugin.PluginContext

//nolint:unused,deadcode
func PluginInit(plugCtx *plugin.PluginContext) {
	pluginContext = plugCtx

	pluginContext.On(plugin.EventTypeMessageCreated, func(m interface{}) {
		fmt.Println("new message", m.(models.Message).Data)
	})

	pluginContext.On(plugin.EventTypeProhibitedStatusChanged, func(s interface{}) {
		fmt.Println("prohibited status changed to", s.(bool))
	})

	r := gin.Default()

	r.GET("/hello/world", helloWorld)
	r.GET("/hello/api/echo/:name", echo)

	pluginContext.RegisterHTTPHandler("hello", r) //nolint:errcheck

	group := plugin.Shortcut.FindGroup("plugins")
	group.CreateItem("hello-simple", "Hello Simple", "测试页面", "/shortcuts/plugin-default.png", "/hello", 99)
}

func echo(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello " + c.Param("name") + ", hello " + pluginContext.ConfigMustGet("hello").(string),
	})
}

func helloWorld(c *gin.Context) {
	currentUser := middlewares.CurrentUser(c.Request)
	c.String(http.StatusOK, "Hello %s", currentUser.FullName)
}
