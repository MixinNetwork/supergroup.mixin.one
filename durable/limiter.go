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
	if !config.AppConfig.System.LimitMessageFrequency {
		return true
	}
	if limiters[key] == nil {
		limiters[key] = rate.NewLimiter(rate.Every(3*time.Minute), 1)
	}
	return limiters[key].Allow()
}
