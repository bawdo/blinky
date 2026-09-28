// Package sticktest provides in-memory fakes of stick.Controller and
// stick.Stick. Nothing here touches hardware or EEPROM.
package sticktest

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/bawdo/go-blinkstick"

	"github.com/bawdo/blinky/internal/stick"
)

// Nano returns the Info of an attached Nano. name may be "".
func Nano(serial, name string) stick.Info {
	return stick.Info{Serial: serial, Firmware: "3.0", Manufacturer: "Agile Innovative Ltd",
		Product: "BlinkStick Nano", Model: "Nano", LEDs: 2, Name: name, Status: stick.StatusOK}
}

// Square returns the Info of an attached Square. name may be "".
func Square(serial, name string) stick.Info {
	return stick.Info{Serial: serial, Firmware: "3.1", Manufacturer: "Agile Innovative Ltd",
		Product: "BlinkStick Square", Model: "Square", LEDs: 8, Name: name, Status: stick.StatusOK}
}

// Busy marks info as held by another process. Its name cannot be read.
func Busy(info stick.Info) stick.Info {
	info.Status = stick.StatusBusy
	info.Name = ""
	return info
}

// Unsupported returns the Info of a model blinky cannot drive.
func Unsupported(serial string) stick.Info {
	return stick.Info{Serial: serial, Firmware: "2.0", Model: "unknown", Status: stick.StatusUnsupported}
}

// Controller is a fake stick.Controller.
type Controller struct {
	mu      sync.Mutex
	infos   []stick.Info
	sticks  map[string]*Stick
	openErr map[string]error
}

// New returns a Controller with infos attached. Every stick with status ok
// gets a fake Stick.
func New(infos ...stick.Info) *Controller {
	c := &Controller{infos: slices.Clone(infos), sticks: map[string]*Stick{}, openErr: map[string]error{}}
	stick.SortBySerial(c.infos)
	for _, info := range c.infos {
		if info.Status == stick.StatusOK {
			c.sticks[info.Serial] = newStick(info)
		}
	}
	return c
}

// Stick returns the fake behind serial, or nil.
func (c *Controller) Stick(serial string) *Stick {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sticks[serial]
}

// FailOpen makes Open(serial) return err.
func (c *Controller) FailOpen(serial string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.openErr[serial] = err
}

// List returns the attached sticks, with names as currently stored.
func (c *Controller) List() ([]stick.Info, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := slices.Clone(c.infos)
	for i, info := range out {
		if s, ok := c.sticks[info.Serial]; ok {
			out[i].Name = s.StoredName()
		}
	}
	return out, nil
}

// Open opens the fake with the given serial.
func (c *Controller) Open(serial string) (stick.Stick, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.openErr[serial]; err != nil {
		return nil, err
	}
	s, ok := c.sticks[serial]
	if !ok {
		return nil, fmt.Errorf("%w: %s", blinkstick.ErrNotFound, serial)
	}
	s.mu.Lock()
	s.closed = false
	s.mu.Unlock()
	return s, nil
}

// Stick is a fake stick.Stick. It scales by the brightness limit as the
// library does, and records every frame and effect call.
type Stick struct {
	mu      sync.Mutex
	info    stick.Info
	leds    []blinkstick.RGB
	frames  [][]blinkstick.RGB
	calls   []string
	blocks  [2][]byte // info blocks 1 and 2; block 1 holds the name
	limit   uint8
	inverse bool
	closed  bool
	err     error
	block   bool
}

func newStick(info stick.Info) *Stick {
	s := &Stick{info: info, leds: make([]blinkstick.RGB, info.LEDs), limit: 255}
	s.blocks[0] = []byte(info.Name)
	return s
}

// FailWith makes every later call that can fail return err. nil clears it.
func (s *Stick) FailWith(err error) { s.mu.Lock(); s.err = err; s.mu.Unlock() }

// BlockEffects makes Blink, Pulse and Morph wait until their context ends.
func (s *Stick) BlockEffects() { s.mu.Lock(); s.block = true; s.mu.Unlock() }

// Current returns the LEDs as stored.
func (s *Stick) Current() []blinkstick.RGB {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.leds)
}

// Frames returns every frame written, oldest first.
func (s *Stick) Frames() [][]blinkstick.RGB {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([][]blinkstick.RGB, len(s.frames))
	for i, f := range s.frames {
		out[i] = slices.Clone(f)
	}
	return out
}

// Calls returns the effect calls made, such as "blink #ff0000 1s 1".
func (s *Stick) Calls() []string { s.mu.Lock(); defer s.mu.Unlock(); return slices.Clone(s.calls) }

// Limit returns the brightness limit last set.
func (s *Stick) Limit() uint8 { s.mu.Lock(); defer s.mu.Unlock(); return s.limit }

// Inverse returns the inverse setting last set.
func (s *Stick) Inverse() bool { s.mu.Lock(); defer s.mu.Unlock(); return s.inverse }

// Closed reports whether the stick is closed.
func (s *Stick) Closed() bool { s.mu.Lock(); defer s.mu.Unlock(); return s.closed }

// Stored returns info block n (1 or 2) without trailing NUL bytes.
func (s *Stick) Stored(n int) []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return bytes.TrimRight(slices.Clone(s.blocks[n-1]), "\x00")
}

// StoredName returns the name held in info block 1.
func (s *Stick) StoredName() string { return string(s.Stored(1)) }

// checkLocked returns the error a call should fail with. s.mu must be held.
func (s *Stick) checkLocked() error {
	if s.closed {
		return blinkstick.ErrClosed
	}
	return s.err
}

func (s *Stick) LEDs() int { return s.info.LEDs }

func (s *Stick) SetFrame(leds []blinkstick.RGB) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.checkLocked(); err != nil {
		return err
	}
	if len(leds) != s.info.LEDs {
		return fmt.Errorf("%w: frame has %d LEDs, want %d", blinkstick.ErrOutOfRange, len(leds), s.info.LEDs)
	}
	next := make([]blinkstick.RGB, len(leds))
	for i, c := range leds {
		next[i] = scale(c, s.limit)
	}
	s.storeLocked(next)
	return nil
}

func (s *Stick) storeLocked(leds []blinkstick.RGB) {
	s.leds = leds
	s.frames = append(s.frames, slices.Clone(leds))
}

func scale(c blinkstick.RGB, limit uint8) blinkstick.RGB {
	f := func(v uint8) uint8 {
		return uint8(uint16(v) * uint16(limit) / 255) //nolint:gosec // G115: the product over 255 is at most 255
	}
	return blinkstick.RGB{R: f(c.R), G: f(c.G), B: f(c.B)}
}

func (s *Stick) Frame() ([]blinkstick.RGB, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.checkLocked(); err != nil {
		return nil, err
	}
	return slices.Clone(s.leds), nil
}

func (s *Stick) checkIndexLocked(i int) error {
	if i < 0 || i >= s.info.LEDs {
		return fmt.Errorf("%w: LED %d, %s has %d", blinkstick.ErrOutOfRange, i, s.info.Model, s.info.LEDs)
	}
	return nil
}

func (s *Stick) SetLED(i int, c blinkstick.RGB) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.checkLocked(); err != nil {
		return err
	}
	if err := s.checkIndexLocked(i); err != nil {
		return err
	}
	next := slices.Clone(s.leds)
	next[i] = scale(c, s.limit)
	s.storeLocked(next)
	return nil
}

func (s *Stick) LED(i int) (blinkstick.RGB, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.checkLocked(); err != nil {
		return blinkstick.RGB{}, err
	}
	if err := s.checkIndexLocked(i); err != nil {
		return blinkstick.RGB{}, err
	}
	return s.leds[i], nil
}

func (s *Stick) Off() error { return s.SetFrame(make([]blinkstick.RGB, s.info.LEDs)) }

func (s *Stick) Blink(ctx context.Context, c blinkstick.RGB, period time.Duration, repeats int) error {
	return s.effect(ctx, fmt.Sprintf("blink %s %v %d", c.Hex(), period, repeats), blinkstick.Off)
}

func (s *Stick) Pulse(ctx context.Context, c blinkstick.RGB, duration time.Duration, repeats int) error {
	return s.effect(ctx, fmt.Sprintf("pulse %s %v %d", c.Hex(), duration, repeats), blinkstick.Off)
}

func (s *Stick) Morph(ctx context.Context, to blinkstick.RGB, duration time.Duration) error {
	return s.effect(ctx, fmt.Sprintf("morph %s %v", to.Hex(), duration), to)
}

// effect records call, then either waits for ctx (BlockEffects) or ends
// with every LED at end, as the library's effects do.
func (s *Stick) effect(ctx context.Context, call string, end blinkstick.RGB) error {
	s.mu.Lock()
	if err := s.checkLocked(); err != nil {
		s.mu.Unlock()
		return err
	}
	s.calls = append(s.calls, call)
	block := s.block
	s.mu.Unlock()
	if block {
		<-ctx.Done()
		return ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	leds := make([]blinkstick.RGB, s.info.LEDs)
	for i := range leds {
		leds[i] = end
	}
	return s.SetFrame(leds)
}

func (s *Stick) Name() (string, error) {
	b, err := s.InfoBlock(1)
	return string(b), err
}

func (s *Stick) SetName(name string) error {
	if len(name) > 32 {
		return fmt.Errorf("%w: name is %d bytes", blinkstick.ErrOutOfRange, len(name))
	}
	return s.SetInfoBlock(1, []byte(name))
}

func (s *Stick) InfoBlock(n int) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.checkLocked(); err != nil {
		return nil, err
	}
	if n != 1 && n != 2 {
		return nil, fmt.Errorf("%w: info block %d", blinkstick.ErrOutOfRange, n)
	}
	return bytes.TrimRight(slices.Clone(s.blocks[n-1]), "\x00"), nil
}

func (s *Stick) SetInfoBlock(n int, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.checkLocked(); err != nil {
		return err
	}
	if n != 1 && n != 2 {
		return fmt.Errorf("%w: info block %d", blinkstick.ErrOutOfRange, n)
	}
	if len(data) > 32 {
		return fmt.Errorf("%w: %d bytes", blinkstick.ErrOutOfRange, len(data))
	}
	s.blocks[n-1] = slices.Clone(data)
	return nil
}

func (s *Stick) SetBrightnessLimit(limit uint8) { s.mu.Lock(); s.limit = limit; s.mu.Unlock() }

func (s *Stick) SetInverse(on bool) { s.mu.Lock(); s.inverse = on; s.mu.Unlock() }

func (s *Stick) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return blinkstick.ErrClosed
	}
	s.closed = true
	return nil
}

var _ stick.Stick = (*Stick)(nil)
