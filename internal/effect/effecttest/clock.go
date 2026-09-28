// Package effecttest provides a manual clock so effects can be tested
// without waiting. Like effect, it imports only the standard library.
package effecttest

import (
	"context"
	"sync"
	"time"
)

// Clock is a fake effect.Clock. Sleep moves time forward at once.
type Clock struct {
	mu         sync.Mutex
	start, now time.Time
}

// NewClock returns a Clock starting at the Unix epoch.
func NewClock() *Clock {
	t := time.Unix(0, 0)
	return &Clock{start: t, now: t}
}

// Now returns the current fake time.
func (c *Clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

// Sleep advances the clock by d, unless ctx is already done.
func (c *Clock) Sleep(ctx context.Context, d time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
	return nil
}

// Elapsed returns how far the clock has moved since NewClock.
func (c *Clock) Elapsed() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now.Sub(c.start)
}
