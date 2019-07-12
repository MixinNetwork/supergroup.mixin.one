package main

import (
	"fmt"

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

	pluginContext.RegisterHTTPHandler("hello", r) //nolint:errcheck

	group := plugin.Shortcut.FindGroup("plugins")
	group.CreateItem("hello0", "Hello 0", "测试0", "", "https://mixin.one", 99)
	group.CreateItem("hello1", "Hello 1", "测试1", "", "https://mixin.one", 0)
}

func helloWorld(c *gin.Context) {
	currentUser := middlewares.CurrentUser(c.Request)
	c.JSON(200, gin.H{
		pluginContext.ConfigMustGet("hello").(string): "Hello, " + currentUser.FullName,
	})
}
