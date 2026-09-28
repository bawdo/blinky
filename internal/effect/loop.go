package effect

import (
	"fmt"
	"time"

	"github.com/bawdo/go-blinkstick"
)

// Loop fades each stick from its starting frame to a colour over fade,
// then back over fade, and repeats. Frames and colours are per stick,
// indexed by sink.
type Loop struct {
	fade time.Duration
	from [][]blinkstick.RGB
	to   []blinkstick.RGB
}

// NewLoop returns a Loop. Each from frame holds one colour per LED and
// repeats to fit a stick with more LEDs; an empty one starts dark.
func NewLoop(fade time.Duration, from [][]blinkstick.RGB, to []blinkstick.RGB) (*Loop, error) {
	switch {
	case fade <= 0:
		return nil, fmt.Errorf("loop: fade %v must be more than 0", fade)
	case len(from) != len(to):
		return nil, fmt.Errorf("loop: %d starting frames for %d sticks", len(from), len(to))
	}
	return &Loop{fade: fade, from: from, to: to}, nil
}

// Cycle is one round trip: there and back.
func (l *Loop) Cycle() time.Duration { return 2 * l.fade }

// Frame implements Effect.
func (l *Loop) Frame(sink, leds int, t time.Duration) []blinkstick.RGB {
	phase := t % l.Cycle()
	level := ramp(phase, l.fade)
	if phase >= l.fade {
		level = 255 - ramp(phase-l.fade, l.fade)
	}
	from := l.from[sink]
	out := make([]blinkstick.RGB, leds)
	for i := range out {
		var f blinkstick.RGB
		if len(from) > 0 {
			f = from[i%len(from)]
		}
		out[i] = blend(f, l.to[sink], level)
	}
	return out
}

// blend moves a towards b by level/255.
func blend(a, b blinkstick.RGB, level uint8) blinkstick.RGB {
	f := func(x, y uint8) uint8 {
		return uint8(int(x) + (int(y)-int(x))*int(level)/255) //nolint:gosec // G115: stays between x and y
	}
	return blinkstick.RGB{R: f(a.R, b.R), G: f(a.G, b.G), B: f(a.B, b.B)}
}
