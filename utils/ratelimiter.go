package utils

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*ipLimiter
	rate     rate.Limit
	burst    int
	ttl      time.Duration
}


func NewIPRateLimiter(limitPerWindow int, window time.Duration) *IPRateLimiter {
	refill := rate.Every(window / time.Duration(limitPerWindow))
	rl := &IPRateLimiter{
		visitors: make(map[string]*ipLimiter),
		rate:     refill,
		burst:    limitPerWindow,
		ttl:      2 * window,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *IPRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIPFromRequest(r)
		if ip == "" {
			next.ServeHTTP(w, r)
			return
		}
		if !rl.getLimiter(ip).Allow() {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if v, ok := rl.visitors[ip]; ok {
		v.lastSeen = time.Now()
		return v.limiter
	}
	l := rate.NewLimiter(rl.rate, rl.burst)
	rl.visitors[ip] = &ipLimiter{limiter: l, lastSeen: time.Now()}
	return l
}

func (rl *IPRateLimiter) cleanupLoop() {
	t := time.NewTicker(rl.ttl / 2)
	defer t.Stop()
	for range t.C {
		now := time.Now()
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if now.Sub(v.lastSeen) > rl.ttl {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func clientIPFromRequest(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			if ip := strings.TrimSpace(parts[0]); ip != "" {
				return ip
			}
		}
	}
	if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
		return xrip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
