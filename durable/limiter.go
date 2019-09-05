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
	duration := config.AppConfig.System.LimitMessageDuration
	if duration <= 0 {
		return true
	}
	if limiters[key] == nil {
		limiters[key] = rate.NewLimiter(rate.Every(time.Duration(duration)*time.Second), 1)
	}
	return limiters[key].Allow()
}
