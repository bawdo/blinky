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

// MorphOptions controls morph.
type MorphOptions struct {
	From     *colour.Spec  // nil means start from what the LEDs show
	Fade     time.Duration // one fade
	Loop     bool          // fade back and forth
	Repeats  int           // round trips with Loop; 0 means until stopped
	Duration time.Duration // stop after this long; 0 means no limit
}

func (m MorphOptions) check() error {
	switch {
	case m.Fade < 0:
		return exitcode.Invalid("--fade must not be negative")
	case m.Repeats < 0:
		return exitcode.Invalid("--repeats must not be negative")
	case m.Duration < 0:
		return exitcode.Invalid("--duration must not be negative")
	case m.Repeats > 0 && !m.Loop:
		return exitcode.Invalid("--repeats needs --loop")
	case m.Loop && m.Fade == 0:
		return exitcode.Invalid("--fade must be more than 0 with --loop")
	}
	return nil
}

// Morph fades the chosen sticks to spec over m.Fade, from m.From if set.
// Without Loop the LEDs stay on spec; with Loop they fade back and forth
// until stopped, then turn off.
func (a *App) Morph(ctx context.Context, req target.Request, set settings.Settings, spec colour.Spec, m MorphOptions) error {
	if err := m.check(); err != nil {
		return err
	}
	infos, _, err := a.resolve(req, target.Group)
	if err != nil {
		return err
	}
	to := pickEach(spec, len(infos))
	var from []blinkstick.RGB
	if m.From != nil {
		from = pickEach(*m.From, len(infos))
	}
	if m.Loop {
		return a.loop(ctx, infos, set, from, to, m)
	}
	if m.Duration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, m.Duration)
		defer cancel()
	}
	return a.each(infos, &set, func(o opened) error {
		if from != nil {
			if err := o.st.SetFrame(fill(o.st.LEDs(), from[o.pos])); err != nil {
				return libError(err)
			}
		}
		return finish(o.st, o.st.Morph(ctx, to[o.pos], m.Fade))
	})
}

// loop fades each stick between its starting frame and to until stopped.
func (a *App) loop(ctx context.Context, infos []stick.Info, set settings.Settings, from, to []blinkstick.RGB, m MorphOptions) error {
	return a.render(ctx, infos, set, limit(m.Repeats, 2*m.Fade, m.Duration), func(sticks []opened) (effect.Effect, error) {
		frames := make([][]blinkstick.RGB, len(infos))
		for _, o := range sticks {
			frames[o.pos] = startFrame(o, from, set)
		}
		return effect.NewLoop(m.Fade, frames, to)
	})
}

// startFrame is the frame a loop starts from on o: the --from-colour if
// given, otherwise what o shows. What o shows is already dimmed by the
// brightness limit, so it is undimmed here to stop the runner dimming it
// again. A stick that cannot be read starts dark.
func startFrame(o opened, from []blinkstick.RGB, set settings.Settings) []blinkstick.RGB {
	if from != nil {
		return []blinkstick.RGB{from[o.pos]}
	}
	shown, err := o.st.Frame()
	if err != nil {
		return nil
	}
	lim := set.Limit()
	for i, c := range shown {
		shown[i] = unscale(c, lim)
	}
	return shown
}

// unscale reverses a brightness limit: the smallest colour that lim dims
// to c. A limit of 0 gives black.
func unscale(c blinkstick.RGB, lim uint8) blinkstick.RGB {
	if lim == 0 {
		return blinkstick.RGB{}
	}
	f := func(v uint8) uint8 {
		return uint8(min((int(v)*255+int(lim)-1)/int(lim), 255)) //nolint:gosec // G115: clamped to 255
	}
	return blinkstick.RGB{R: f(c.R), G: f(c.G), B: f(c.B)}
}

// fill returns n LEDs of colour c.
func fill(n int, c blinkstick.RGB) []blinkstick.RGB {
	out := make([]blinkstick.RGB, n)
	for i := range out {
		out[i] = c
	}
	return out
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
