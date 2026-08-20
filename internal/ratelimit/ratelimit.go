package ratelimit

import (
	"sync"
	"time"

	"github.com/lacsar712/corebore/internal/clock"
)

// Bucket is a token bucket. Take returns how long the caller should wait
// if no token is available (0 means proceed now).
type Bucket struct {
	mu       sync.Mutex
	clk      clock.Clock
	rate     float64
	burst    float64
	tokens   float64
	lastTime time.Time
}

func New(clk clock.Clock, ratePerSec float64, burst int) *Bucket {
	if ratePerSec <= 0 {
		ratePerSec = 5
	}
	if burst < 1 {
		burst = 5
	}
	now := clk.Now()
	return &Bucket{
		clk:      clk,
		rate:     ratePerSec,
		burst:    float64(burst),
		tokens:   float64(burst),
		lastTime: now,
	}
}

// Take consumes one token if one is available now and returns 0 (proceed).
// If no token is available it leaves the bucket untouched and returns how
// long the caller should wait before the next token refills. The caller is
// expected to retry after that delay rather than sleep in place, so Take does
// not reserve future tokens — a token is only debited when it returns 0.
func (b *Bucket) Take() (wait time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.clk.Now()
	if elapsed := now.Sub(b.lastTime); elapsed > 0 {
		b.tokens += b.rate * elapsed.Seconds()
		if b.tokens > b.burst {
			b.tokens = b.burst
		}
		b.lastTime = now
	}
	if b.tokens >= 1 {
		b.tokens--
		return 0
	}
	deficit := 1 - b.tokens
	return time.Duration((deficit / b.rate) * float64(time.Second))
}

type Snapshot struct {
	Rate     float64   `json:"rate"`
	Burst    float64   `json:"burst"`
	Tokens   float64   `json:"tokens"`
	LastTime time.Time `json:"last_time"`
}

func (b *Bucket) Snapshot() Snapshot {
	b.mu.Lock()
	defer b.mu.Unlock()
	return Snapshot{Rate: b.rate, Burst: b.burst, Tokens: b.tokens, LastTime: b.lastTime}
}

func (b *Bucket) Restore(s Snapshot) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if s.Rate > 0 {
		b.rate = s.Rate
	}
	if s.Burst > 0 {
		b.burst = s.Burst
	}
	b.tokens = s.Tokens
	b.lastTime = s.LastTime
}
