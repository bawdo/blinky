package effect

import (
	"fmt"
	"time"

	"github.com/bawdo/go-blinkstick"
)

// Blink lights every LED of a stick for on, then leaves it dark for off.
// With a second colour, each cycle blinks the first colour then the
// second. Colours are per stick, indexed by sink.
type Blink struct {
	on, off       time.Duration
	first, second []blinkstick.RGB
}

// NewBlink returns a Blink. second is nil for a single colour, or holds
// one colour per stick like first.
func NewBlink(on, off time.Duration, first, second []blinkstick.RGB) (*Blink, error) {
	switch {
	case on <= 0:
		return nil, fmt.Errorf("blink: on time %v must be more than 0", on)
	case off < 0:
		return nil, fmt.Errorf("blink: off time %v must not be negative", off)
	case second != nil && len(second) != len(first):
		return nil, fmt.Errorf("blink: %d second colours for %d sticks", len(second), len(first))
	}
	return &Blink{on: on, off: off, first: first, second: second}, nil
}

// Cycle is one repeat: on then off, twice over with a second colour.
func (b *Blink) Cycle() time.Duration {
	c := b.on + b.off
	if b.second != nil {
		c *= 2
	}
	return c
}

// Frame implements Effect.
func (b *Blink) Frame(sink, leds int, t time.Duration) []blinkstick.RGB {
	phase := t % b.Cycle()
	colours := b.first
	if half := b.on + b.off; phase >= half {
		phase -= half
		colours = b.second
	}
	out := make([]blinkstick.RGB, leds)
	if phase < b.on {
		for i := range out {
			out[i] = colours[sink]
		}
	}
	return out
}
