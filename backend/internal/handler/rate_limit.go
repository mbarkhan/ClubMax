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
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}
			now := time.Now()
			mu.Lock()
			entry := attempts[host]
			if now.After(entry.until) {
				entry = attemptWindow{until: now.Add(window)}
			}
			entry.count++
			attempts[host] = entry
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
