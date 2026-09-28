package app

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/bawdo/go-blinkstick"

	"github.com/bawdo/blinky/internal/colour"
	"github.com/bawdo/blinky/internal/effect"
	"github.com/bawdo/blinky/internal/exitcode"
	"github.com/bawdo/blinky/internal/render"
	"github.com/bawdo/blinky/internal/settings"
	"github.com/bawdo/blinky/internal/target"
)

// DiscoOptions controls disco.
type DiscoOptions struct {
	MinPeriod, MaxPeriod time.Duration
	MaxGap               time.Duration
	Duration             time.Duration // 0 means until stopped
	Palette              []colour.Spec // empty means vivid
	Seed                 uint64        // 0 means random
}

// PoliceOptions controls police.
type PoliceOptions struct {
	Period    time.Duration
	Duration  time.Duration // 0 means until stopped
	Alternate bool
}

// Disco pulses every LED in random colours until stopped.
func (a *App) Disco(ctx context.Context, req target.Request, set settings.Settings, o DiscoOptions) error {
	palette := o.Palette
	if len(palette) == 0 {
		palette = []colour.Spec{{Kind: colour.Vivid}}
	}
	d, err := effect.NewDisco(effect.DiscoConfig{
		MinPeriod: o.MinPeriod,
		MaxPeriod: o.MaxPeriod,
		MaxGap:    o.MaxGap,
		Rand:      newRand(o.Seed),
		Colour: func(r *rand.Rand) blinkstick.RGB {
			return palette[r.IntN(len(palette))].Pick(r)
		},
	})
	if err != nil {
		return fmt.Errorf("%w: %w", exitcode.ErrInvalidArgs, err)
	}
	return a.run(ctx, req, set, d, o.Duration)
}

// Police crossfades red and blue halves until stopped.
func (a *App) Police(ctx context.Context, req target.Request, set settings.Settings, o PoliceOptions) error {
	p, err := effect.NewPolice(o.Period, o.Alternate)
	if err != nil {
		return fmt.Errorf("%w: %w", exitcode.ErrInvalidArgs, err)
	}
	return a.run(ctx, req, set, p, o.Duration)
}

// run renders e on every chosen stick until ctx ends or duration passes,
// then turns the LEDs off.
func (a *App) run(ctx context.Context, req target.Request, set settings.Settings, e effect.Effect, duration time.Duration) error {
	if duration < 0 {
		return exitcode.Invalid("--duration must not be negative")
	}
	infos, _, err := a.resolve(req, target.Group)
	if err != nil {
		return err
	}
	return a.session(infos, &set, func(sticks []opened) []error {
		sinks := make([]effect.Sink, len(sticks))
		for j, o := range sticks {
			sinks[j] = o.st
		}
		runner := effect.Runner{Duration: duration, Clock: a.clock, OnEvent: func(ev effect.Event) {
			id := render.Sanitise(sticks[ev.Sink].info.ID())
			switch ev.Kind {
			case effect.Dropped:
				_, _ = fmt.Fprintf(a.err, "%s: disconnected, retrying\n", id)
			case effect.Returned:
				_, _ = fmt.Fprintf(a.err, "%s: reconnected\n", id)
			}
		}}
		errs := runner.Run(ctx, e, sinks)
		for j, o := range sticks {
			if errs[j] != nil {
				errs[j] = libError(errs[j])
				continue
			}
			errs[j] = libError(o.st.Off())
		}
		return errs
	})
}
