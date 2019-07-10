package main

import (
	"fmt"

	"github.com/MixinNetwork/supergroup.mixin.one/models"
	"github.com/MixinNetwork/supergroup.mixin.one/plugin"
	"github.com/gin-gonic/gin"
)

//nolint:unused,deadcode
func PluginInit(ctx *plugin.PluginContext) {
	ctx.On(plugin.EventTypeMessageCreated, func(m interface{}) {
		fmt.Println("new message", m.(models.Message).Data)
	})

	ctx.On(plugin.EventTypeProhibitedStatusChanged, func(s interface{}) {
		fmt.Println("prohibited status changed to", s.(bool))
	})

	r := gin.Default()

	r.GET("/hello/world", helloWorld)

	ctx.RegisterHTTPHandler("hello", r) //nolint:errcheck
}

func helloWorld(c *gin.Context) {
	c.JSON(200, gin.H{
		"hello": "Hello, " + pluginContext.ConfigMustGet("hello").(string),
	})
}
