package util

import (
	"sync"
	"time"
)

type RateLimiter struct {
	limits map[string]*requestWindow
	mu     sync.Mutex
}

type requestWindow struct {
	count   int
	earliest time.Time
}

var (
	rateLimiterInstance = &RateLimiter{limits: make(map[string]*requestWindow)}
	maxRequests         = 5
	timeWindow          = 10 * time.Second
)

func AllowRequest(key string) bool {
	rateLimiterInstance.mu.Lock()
	defer rateLimiterInstance.mu.Unlock()

	now := time.Now()
	window, exists := rateLimiterInstance.limits[key]
	if !exists || now.Sub(window.earliest) >= timeWindow {
		rateLimiterInstance.limits[key] = &requestWindow{count: 1, earliest: now}
		return true
	}

	if window.count < maxRequests {
		window.count++
		return true
	}

	return false
}
