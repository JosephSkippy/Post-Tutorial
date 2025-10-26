package ratelimiter

import (
	"sync"
	"time"
)

type clientData struct {
	count     int
	startTime time.Time
}

type FixedWindowRateLimiter struct {
	sync.RWMutex
	clients map[string]*clientData
	limit   int
	window  time.Duration
}

func NewFixedWindowLimiter(limit int, window time.Duration) *FixedWindowRateLimiter {
	rl := &FixedWindowRateLimiter{
		clients: make(map[string]*clientData),
		limit:   limit,
		window:  window,
	}

	go rl.cleanupLoop() // ✅ only ONE cleanup goroutine
	return rl
}

func (rl *FixedWindowRateLimiter) Allow(ip string) (bool, time.Duration) {
	rl.Lock()
	defer rl.Unlock()

	now := time.Now()
	data, exists := rl.clients[ip]

	// First request or window expired → start new window
	if !exists || now.Sub(data.startTime) >= rl.window {
		rl.clients[ip] = &clientData{
			count:     1,
			startTime: now,
		}
		return true, 0
	}

	// Within window and still under limit → allow
	if data.count < rl.limit {
		data.count++
		return true, 0
	}

	// Over the limit → reject
	wait := rl.window - now.Sub(data.startTime)
	return false, wait
}

// cleanupLoop periodically removes expired windows
func (rl *FixedWindowRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.window)
	for range ticker.C {
		now := time.Now()

		rl.Lock()
		for ip, data := range rl.clients {
			if now.Sub(data.startTime) >= rl.window {
				delete(rl.clients, ip)
			}
		}
		rl.Unlock()
	}
}
