package handler

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type attemptWindow struct {
	count int
	until time.Time
}

func rateLimit(max int, window time.Duration) func(http.Handler) http.Handler {
	var mu sync.Mutex
	attempts := make(map[string]attemptWindow)
	lastSweep := time.Now()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}
			key := "ip:" + host
			if actor := current(r); actor != nil {
				key = "user:" + actor.ID
			}
			now := time.Now()
			mu.Lock()
			if now.Sub(lastSweep) >= window {
				for candidate, entry := range attempts {
					if now.After(entry.until) {
						delete(attempts, candidate)
					}
				}
				lastSweep = now
			}
			entry := attempts[key]
			if now.After(entry.until) {
				entry = attemptWindow{until: now.Add(window)}
			}
			entry.count++
			attempts[key] = entry
			blocked := entry.count > max
			mu.Unlock()
			if blocked {
				w.Header().Set("Retry-After", "60")
				write(w, http.StatusTooManyRequests, map[string]string{"error": "rate_limit", "message": "تعداد درخواست‌ها بیش از حد مجاز است؛ کمی بعد دوباره تلاش کنید."})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
