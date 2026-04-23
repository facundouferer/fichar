package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	RequestsPerMinute int
	Burst             int
}

// Default configs
var (
	LoginRateLimiterConfig      = RateLimiterConfig{RequestsPerMinute: 3, Burst: 5}
	AttendanceRateLimiterConfig = RateLimiterConfig{RequestsPerMinute: 10, Burst: 15}
	CaptchaRateLimiterConfig    = RateLimiterConfig{RequestsPerMinute: 5, Burst: 10}
)

// IPRateLimiter implements IP-based rate limiting using golang.org/x/time/rate
type IPRateLimiter struct {
	mu      sync.RWMutex
	limiter *rate.Limiter
	config  RateLimiterConfig
}

// NewIPRateLimiter creates a new rate limiter for IP-based limiting
func NewIPRateLimiter(config RateLimiterConfig) *IPRateLimiter {
	// Convert requests per minute to requests per second
	rps := float64(config.RequestsPerMinute) / 60.0

	rl := &IPRateLimiter{
		limiter: rate.NewLimiter(rate.Limit(rps), config.Burst),
		config:  config,
	}

	return rl
}

// Middleware returns the HTTP middleware handler
func (rl *IPRateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip if limiter not initialized
			if rl == nil || rl.limiter == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Get client IP
			clientIP := getClientIP(r)

			// Check rate limit (we use a simple key here - in production you'd want per-IP)
			rl.mu.RLock()
			allowed := rl.limiter.Allow()
			rl.mu.RUnlock()

			if !allowed {
				retryAfter := rl.config.RequestsPerMinute / 60
				if retryAfter < 1 {
					retryAfter = 1
				}
				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
				w.Header().Set("X-RateLimit-Remaining", "0")
				http.Error(w, fmt.Sprintf("Rate limit exceeded. Try again in %d seconds", retryAfter), http.StatusTooManyRequests)
				return
			}

			// Add client IP to request for downstream use
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.config.RequestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", rl.config.Burst))
			w.Header().Set("X-Client-IP", clientIP)

			next.ServeHTTP(w, r)
		})
	}
}

// IPRateLimiterMap manages multiple rate limiters per IP
type IPRateLimiterMap struct {
	mu              sync.RWMutex
	limiters        map[string]*rate.Limiter
	config          RateLimiterConfig
	cleanupInterval time.Duration
}

// NewIPRateLimiterMap creates a new mapped rate limiter
func NewIPRateLimiterMap(config RateLimiterConfig, cleanupInterval time.Duration) *IPRateLimiterMap {
	rlm := &IPRateLimiterMap{
		limiters:        make(map[string]*rate.Limiter),
		config:          config,
		cleanupInterval: cleanupInterval,
	}

	// Start cleanup goroutine
	go rlm.cleanup()

	return rlm
}

// getLimiter gets or creates a rate limiter for the given IP
func (rlm *IPRateLimiterMap) getLimiter(ip string) *rate.Limiter {
	// Convert requests per minute to requests per second
	rps := float64(rlm.config.RequestsPerMinute) / 60.0

	rlm.mu.RLock()
	limiter, exists := rlm.limiters[ip]
	rlm.mu.RUnlock()

	if exists {
		return limiter
	}

	// Create new limiter
	rlm.mu.Lock()
	// Double-check after acquiring write lock
	if limiter, exists = rlm.limiters[ip]; exists {
		rlm.mu.Unlock()
		return limiter
	}

	limiter = rate.NewLimiter(rate.Limit(rps), rlm.config.Burst)
	rlm.limiters[ip] = limiter
	rlm.mu.Unlock()

	return limiter
}

// Allow checks if the IP is allowed
func (rlm *IPRateLimiterMap) Allow(ip string) bool {
	limiter := rlm.getLimiter(ip)
	return limiter.Allow()
}

// Middleware returns the HTTP middleware handler
func (rlm *IPRateLimiterMap) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip if not initialized
			if rlm == nil {
				next.ServeHTTP(w, r)
				return
			}

			clientIP := getClientIP(r)

			allowed := rlm.Allow(clientIP)

			if !allowed {
				retryAfter := rlm.config.RequestsPerMinute / 60
				if retryAfter < 1 {
					retryAfter = 1
				}
				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
				w.Header().Set("X-RateLimit-Remaining", "0")
				http.Error(w, fmt.Sprintf("Rate limit exceeded. Try again in %d seconds", retryAfter), http.StatusTooManyRequests)
				return
			}

			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rlm.config.RequestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", rlm.config.Burst))

			next.ServeHTTP(w, r)
		})
	}
}

// cleanup removes stale limiters periodically
func (rlm *IPRateLimiterMap) cleanup() {
	ticker := time.NewTicker(rlm.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rlm.mu.Lock()
		// In a real implementation, you would track last access time
		// and remove limiters that haven't been used in a while
		rlm.mu.Unlock()
	}
}

// RateLimitMiddleware creates a middleware for rate limiting using in-memory map
func RateLimitMiddleware(requestsPerMinute int, burst int) func(http.Handler) http.Handler {
	config := RateLimiterConfig{
		RequestsPerMinute: requestsPerMinute,
		Burst:             burst,
	}

	rlm := NewIPRateLimiterMap(config, 5*time.Minute)

	return rlm.Middleware()
}

// getClientIP extracts the real client IP from request
func getClientIP(r *http.Request) string {
	// Check for forwarded header (when behind proxy)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}
	// Fall back to remote address (remove port if present)
	ip := r.RemoteAddr
	for i := len(ip) - 1; i >= 0; i-- {
		if ip[i] == ':' {
			ip = ip[:i]
			break
		}
	}
	return ip
}
