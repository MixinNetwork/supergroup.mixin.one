package durable

import (
	"time"

	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"golang.org/x/time/rate"
)

var limiters map[string]*rate.Limiter = make(map[string]*rate.Limiter, 0)

func Allow(key string) bool {
	if config.AppConfig.Service.Environment == "test" {
		return true
	}
	system := config.AppConfig.System
	if system.LimitMessageDuration <= 0 || system.LimitMessageNumber <= 0 {
		return true
	}
	if limiters[key] == nil {
		limiters[key] = rate.NewLimiter(rate.Every(time.Duration(system.LimitMessageDuration)*time.Second), system.LimitMessageNumber)
	}
	return limiters[key].Allow()
}
