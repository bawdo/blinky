package app

import (
	"context"
	"fmt"
	"time"

	"github.com/bawdo/go-blinkstick"

	"github.com/bawdo/blinky/internal/colour"
	"github.com/bawdo/blinky/internal/effect"
	"github.com/bawdo/blinky/internal/exitcode"
	"github.com/bawdo/blinky/internal/settings"
	"github.com/bawdo/blinky/internal/stick"
	"github.com/bawdo/blinky/internal/target"
)

// Repeat controls pulse.
type Repeat struct {
	Period   time.Duration // one cycle
	Repeats  int           // cycles to run; 0 means until stopped
	Duration time.Duration // stop after this long; 0 means no limit
}

func (r Repeat) check() error {
	switch {
	case r.Period <= 0:
		return exitcode.Invalid("--period must be more than 0")
	case r.Repeats < 0:
		return exitcode.Invalid("--repeats must not be negative")
	case r.Duration < 0:
		return exitcode.Invalid("--duration must not be negative")
	}
	return nil
}

// BlinkOptions controls blink.
type BlinkOptions struct {
	Period   time.Duration  // one on and off cycle
	On, Off  *time.Duration // nil means half of Period
	Second   *colour.Spec   // nil means one colour
	Repeats  int            // cycles to run; 0 means until stopped
	Duration time.Duration  // stop after this long; 0 means no limit
}

// times checks o and returns how long to stay lit and dark.
func (o BlinkOptions) times() (on, off time.Duration, err error) {
	switch {
	case o.Period <= 0:
		return 0, 0, exitcode.Invalid("--period must be more than 0")
	case o.Repeats < 0:
		return 0, 0, exitcode.Invalid("--repeats must not be negative")
	case o.Duration < 0:
		return 0, 0, exitcode.Invalid("--duration must not be negative")
	case o.On != nil && *o.On <= 0:
		return 0, 0, exitcode.Invalid("--on-time must be more than 0")
	case o.Off != nil && *o.Off < 0:
		return 0, 0, exitcode.Invalid("--off-time must not be negative")
	}
	on, off = o.Period/2, o.Period/2
	if o.On != nil {
		on = *o.On
	}
	if o.Off != nil {
		off = *o.Off
	}
	return on, off, nil
}

// Blink turns the chosen sticks on and off, alternating with a second
// colour if one is given, then leaves them off.
func (a *App) Blink(ctx context.Context, req target.Request, set settings.Settings, spec colour.Spec, o BlinkOptions) error {
	on, off, err := o.times()
	if err != nil {
		return err
	}
	infos, _, err := a.resolve(req, target.Group)
	if err != nil {
		return err
	}
	var second []blinkstick.RGB
	if o.Second != nil {
		second = pickEach(*o.Second, len(infos))
	}
	b, err := effect.NewBlink(on, off, pickEach(spec, len(infos)), second)
	if err != nil {
		return fmt.Errorf("%w: %w", exitcode.ErrInvalidArgs, err)
	}
	return a.render(ctx, infos, set, limit(o.Repeats, b.Cycle(), o.Duration), constant(b))
}

// Pulse fades the chosen sticks up and down.
func (a *App) Pulse(ctx context.Context, req target.Request, set settings.Settings, spec colour.Spec, r Repeat) error {
	return a.repeat(ctx, req, set, spec, r, func(ctx context.Context, st stick.Stick, c blinkstick.RGB) error {
		return st.Pulse(ctx, c, r.Period, 1)
	})
}

// repeat runs once per cycle on every stick at the same time, for pulse.
// The library effects run one cycle per call so --duration can end between
// cycles.
func (a *App) repeat(ctx context.Context, req target.Request, set settings.Settings, spec colour.Spec, r Repeat,
	once func(context.Context, stick.Stick, blinkstick.RGB) error) error {
	if err := r.check(); err != nil {
		return err
	}
	infos, _, err := a.resolve(req, target.Group)
	if err != nil {
		return err
	}
	colours := pickEach(spec, len(infos))
	if r.Duration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.Duration)
		defer cancel()
	}
	return a.each(infos, &set, func(o opened) error {
		for n := 0; r.Repeats == 0 || n < r.Repeats; n++ {
			if err := once(ctx, o.st, colours[o.pos]); err != nil {
				return finish(o.st, err)
			}
		}
		return nil
	})
}

// Morph fades the chosen sticks from their current colours to spec over d.
func (a *App) Morph(ctx context.Context, req target.Request, set settings.Settings, spec colour.Spec, d time.Duration) error {
	if d < 0 {
		return exitcode.Invalid("--duration must not be negative")
	}
	infos, _, err := a.resolve(req, target.Group)
	if err != nil {
		return err
	}
	colours := pickEach(spec, len(infos))
	return a.each(infos, &set, func(o opened) error {
		return finish(o.st, o.st.Morph(ctx, colours[o.pos], d))
	})
}

// pickEach picks spec once per stick, so random and vivid differ per stick.
func pickEach(spec colour.Spec, n int) []blinkstick.RGB {
	r := newRand(0)
	out := make([]blinkstick.RGB, n)
	for i := range out {
		out[i] = spec.Pick(r)
	}
	return out
}
