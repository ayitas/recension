package httpapi

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// authLimiter is a tiny in-memory rate limit for login/signup.
type authLimiter struct {
	mu      sync.Mutex
	hits    map[string][]time.Time
	limit   int
	window  time.Duration
	lastGC  time.Time
}

func newAuthLimiter(limit int, window time.Duration) *authLimiter {
	return &authLimiter{
		hits:   map[string][]time.Time{},
		limit:  limit,
		window: window,
	}
}

func (l *authLimiter) allow(key string) bool {
	now := time.Now()
	cutoff := now.Add(-l.window)

	l.mu.Lock()
	defer l.mu.Unlock()

	if now.Sub(l.lastGC) > l.window {
		for k, times := range l.hits {
			kept := times[:0]
			for _, t := range times {
				if t.After(cutoff) {
					kept = append(kept, t)
				}
			}
			if len(kept) == 0 {
				delete(l.hits, k)
			} else {
				l.hits[k] = kept
			}
		}
		l.lastGC = now
	}

	times := l.hits[key]
	kept := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= l.limit {
		l.hits[key] = kept
		return false
	}
	l.hits[key] = append(kept, now)
	return true
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
