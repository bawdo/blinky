package effect

import (
	"time"

	"github.com/bawdo/go-blinkstick"
)

// ramp returns how far x is through total, from 0 to 255.
func ramp(x, total time.Duration) uint8 {
	if total <= 0 {
		return 255
	}
	v := min(max(int64(x)*255/int64(total), 0), 255)
	return uint8(v) //nolint:gosec // G115: clamped to 0..255 above
}

// dim scales c by level/255.
func dim(c blinkstick.RGB, level uint8) blinkstick.RGB {
	f := func(v uint8) uint8 {
		return uint8(uint16(v) * uint16(level) / 255) //nolint:gosec // G115: the product over 255 is at most 255
	}
	return blinkstick.RGB{R: f(c.R), G: f(c.G), B: f(c.B)}
}
