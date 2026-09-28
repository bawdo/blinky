// Package settings resolves the brightness and inverse settings a command
// applies to every stick it opens.
//
// Settings come from ordered layers, later ones winning. The first release
// has one layer, the command line flags. A config file layer will go in
// front of it, so flags keep precedence.
package settings

import (
	"fmt"

	"github.com/bawdo/blinky/internal/exitcode"
)

// Settings apply to every stick a command opens.
type Settings struct {
	Brightness int  // percent, 0 to 100
	Inverse    bool // flip every colour, for LEDs wired so 255 means off
}

// Layer is one source of settings. A nil field is not set by this layer.
type Layer struct {
	Brightness *int
	Inverse    *bool
}

// Default is full brightness, not inverted.
func Default() Settings {
	return Settings{Brightness: 100}
}

// Resolve starts from Default and applies layers in order.
func Resolve(layers ...Layer) (Settings, error) {
	s := Default()
	for _, l := range layers {
		if l.Brightness != nil {
			s.Brightness = *l.Brightness
		}
		if l.Inverse != nil {
			s.Inverse = *l.Inverse
		}
	}
	if s.Brightness < 0 || s.Brightness > 100 {
		return Settings{}, fmt.Errorf("%w: brightness %d, want 0 to 100", exitcode.ErrInvalidArgs, s.Brightness)
	}
	return s, nil
}

// Limit converts Brightness to go-blinkstick's 0 to 255 brightness limit.
func (s Settings) Limit() uint8 {
	return uint8((s.Brightness*255 + 50) / 100) //nolint:gosec // G115: Resolve keeps Brightness in 0..100
}

// Target is anything settings can be applied to, such as a stick.
type Target interface {
	SetBrightnessLimit(limit uint8)
	SetInverse(on bool)
}

// Apply sets t's brightness limit and inverse mode.
func (s Settings) Apply(t Target) {
	t.SetBrightnessLimit(s.Limit())
	t.SetInverse(s.Inverse)
}
