package main

import (
	"fmt"

	"github.com/MixinNetwork/supergroup.mixin.one/models"
	"github.com/MixinNetwork/supergroup.mixin.one/plugin"
	"github.com/gin-gonic/gin"
)

func init() {
	plugin.On(plugin.EventTypeMessageCreated, func(m interface{}) {
		fmt.Print(m.(models.Message).Data)
	})

	r := gin.Default()

	r.GET("/hello/world", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"hello": "Hello, world!",
		})
	})

	plugin.RegisterHTTPHandler("hello", r)
}
