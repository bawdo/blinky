package effect

import (
	"fmt"
	"time"

	"github.com/bawdo/go-blinkstick"
)

// Police crossfades red and blue between the two halves of every stick.
// In the first half of each period the red half fades up while the blue
// half fades out; in the second half they swap back. Every stick shares
// one clock, so they flash in step.
type Police struct {
	period    time.Duration
	alternate bool
}

// NewPolice returns a Police effect. With alternate, every second stick
// swaps halves so neighbours run in opposite phase.
func NewPolice(period time.Duration, alternate bool) (*Police, error) {
	if period <= 0 {
		return nil, fmt.Errorf("police: period %v must be more than 0", period)
	}
	return &Police{period: period, alternate: alternate}, nil
}

// Frame implements Effect.
func (p *Police) Frame(sink, leds int, t time.Duration) []blinkstick.RGB {
	half := p.period / 2
	phase := t % p.period
	var red, blue uint8
	if phase < half {
		up := ramp(phase, half)
		red, blue = up, 255-up
	} else {
		up := ramp(phase-half, p.period-half)
		red, blue = 255-up, up
	}
	swap := p.alternate && sink%2 == 1
	out := make([]blinkstick.RGB, leds)
	for i := range out {
		if (i*2/leds == 0) != swap {
			out[i] = blinkstick.RGB{R: red}
		} else {
			out[i] = blinkstick.RGB{B: blue}
		}
	}
	return out
}
