// Package stick is the seam between blinky and the hardware. Commands
// reach sticks only through Controller and Stick, so a daemon client can
// replace the direct go-blinkstick implementation without touching them.
package stick

import (
	"context"
	"time"

	"github.com/bawdo/go-blinkstick"
)

// Status says whether blinky can drive an attached stick.
type Status string

const (
	StatusOK          Status = "ok"
	StatusBusy        Status = "busy"        // another process holds it
	StatusUnsupported Status = "unsupported" // a model go-blinkstick cannot drive
)

// Info describes an attached stick.
type Info struct {
	Serial       string
	Firmware     string
	Manufacturer string
	Product      string
	Model        string // "Nano", "Square" or "unknown"
	LEDs         int    // 0 when unsupported
	Name         string // "" if none is set, or if Busy
	Status       Status
}

// ID is the string to pass to --device: the name if set, otherwise the
// serial.
func (i Info) ID() string {
	if i.Name != "" {
		return i.Name
	}
	return i.Serial
}

// Stick is an open BlinkStick. *blinkstick.Device has every method but
// LEDs.
type Stick interface {
	LEDs() int
	SetFrame(leds []blinkstick.RGB) error
	Frame() ([]blinkstick.RGB, error)
	SetLED(i int, c blinkstick.RGB) error
	LED(i int) (blinkstick.RGB, error)
	Off() error
	Blink(ctx context.Context, c blinkstick.RGB, period time.Duration, repeats int) error
	Pulse(ctx context.Context, c blinkstick.RGB, duration time.Duration, repeats int) error
	Morph(ctx context.Context, to blinkstick.RGB, duration time.Duration) error
	Name() (string, error)
	SetName(name string) error
	InfoBlock(n int) ([]byte, error)
	SetInfoBlock(n int, data []byte) error
	SetBrightnessLimit(limit uint8)
	SetInverse(on bool)
	Close() error
}

// Controller finds and opens sticks.
type Controller interface {
	// List returns every attached stick, sorted by serial.
	List() ([]Info, error)
	// Open opens the stick with the given serial.
	Open(serial string) (Stick, error)
}
