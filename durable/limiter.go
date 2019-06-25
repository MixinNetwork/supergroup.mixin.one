package durable

import (
	"time"

	"golang.org/x/time/rate"
)

var limiters map[string]*rate.Limiter = make(map[string]*rate.Limiter, 0)

func Allow(key string) bool {
	if limiters[key] == nil {
		limiters[key] = rate.NewLimiter(rate.Every(30*time.Second), 10)
	}
	return limiters[key].Allow()
}
