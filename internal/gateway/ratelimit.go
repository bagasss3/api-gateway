package gateway

import (
	"sync"
	"time"
)

type tokenBucket struct {
	ratePerSec int
	burst      int
	tokens     float64
	last       time.Time
	mu         sync.Mutex
}

func newBucket(rps, burst int) *tokenBucket {
	return &tokenBucket{ratePerSec: rps, burst: burst, tokens: float64(burst), last: time.Now()}
}

func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * float64(b.ratePerSec)
	if b.tokens > float64(b.burst) {
		b.tokens = float64(b.burst)
	}
	b.last = now
	if b.tokens >= 1 {
		b.tokens -= 1
		return true
	}
	return false
}

type ipLimiter struct {
	perIP  map[string]*tokenBucket
	global *tokenBucket
	mu     sync.Mutex
}

func newIPLimiter(globalRPS int) *ipLimiter {
	return &ipLimiter{perIP: make(map[string]*tokenBucket), global: newBucket(globalRPS, globalRPS)}
}

func (l *ipLimiter) allow(ip string, perIPRPS, perIPBurst int) bool {
	if l.global != nil && !l.global.allow() {
		return false
	}
	l.mu.Lock()
	bkt, ok := l.perIP[ip]
	if !ok {
		bkt = newBucket(perIPRPS, perIPBurst)
		l.perIP[ip] = bkt
	}
	l.mu.Unlock()
	return bkt.allow()
}
