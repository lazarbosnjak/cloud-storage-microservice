package gateway

import (
	"context"
	"math"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"sync"
	"time"

	"api-gateway/internal/startup/config"

	"golang.org/x/time/rate"
)

const maxCleanupInterval = time.Minute * 10

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimitHandler is a token bucket rate limiter with one bucket per client
// IP (IPv6 clients are grouped by /64 prefix).
type IPRateLimitHandler struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rps      rate.Limit
	burst    int
	ttl      time.Duration

	cfg *config.Config
}

func NewRateLimitHandler(cfg *config.Config) *IPRateLimitHandler {
	return &IPRateLimitHandler{
		visitors: make(map[string]*visitor),
		rps:      rate.Limit(cfg.RateLimitRPS),
		burst:    int(cfg.RateLimitBucketSize),
		ttl:      cfg.RateLimitIdleTTL,
		cfg:      cfg,
	}
}

func (h *IPRateLimitHandler) Middleware(next http.Handler, reject http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter := h.getLimiter(h.bucketKey(r))

		res := limiter.Reserve()
		if !res.OK() {
			reject(w, r)
			return
		}

		if delay := res.Delay(); delay > 0 {
			res.Cancel()
			w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(delay.Seconds()))))
			reject(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *IPRateLimitHandler) RunCleanup(ctx context.Context) {
	interval := min(h.ttl, maxCleanupInterval)
	if interval <= 0 {
		interval = maxCleanupInterval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			h.removeIdle(now)
		}
	}
}

func (h *IPRateLimitHandler) removeIdle(now time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for key, v := range h.visitors {
		isIdle := now.Sub(v.lastSeen) > h.ttl
		if isIdle {
			delete(h.visitors, key)
		}
	}
}

func (h *IPRateLimitHandler) getLimiter(key string) *rate.Limiter {
	h.mu.Lock()
	defer h.mu.Unlock()

	v, ok := h.visitors[key]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(h.rps, h.burst)}
		h.visitors[key] = v
	}
	v.lastSeen = time.Now()

	return v.limiter
}

func (h *IPRateLimitHandler) getClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

func (h *IPRateLimitHandler) bucketKey(r *http.Request) string {
	ip := h.getClientIP(r)

	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return ip
	}
	addr = addr.Unmap()
	if addr.Is6() {
		p, _ := addr.Prefix(64)
		return p.String()
	}
	return addr.String()
}
