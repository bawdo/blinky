// Package effect renders animations frame by frame onto one or more
// sticks.
//
// It imports only the standard library and go-blinkstick, so it can move
// into go-blinkstick unchanged apart from the import path. A test enforces
// this.
package effect

import (
	"cmp"
	"context"
	"errors"
	"time"

	"github.com/bawdo/go-blinkstick"
)

// Sink is somewhere frames go. An open stick is one.
type Sink interface {
	LEDs() int
	SetFrame(leds []blinkstick.RGB) error
}

// Effect produces the frame for sink (its index in the Runner's list),
// which has leds LEDs, at time t since the effect started. The Runner calls
// Frame from one goroutine, with t never going backwards for a sink.
type Effect interface {
	Frame(sink, leds int, t time.Duration) []blinkstick.RGB
}

// Clock is the time source a Runner uses. Tests swap in a manual one.
type Clock interface {
	Now() time.Time
	Sleep(ctx context.Context, d time.Duration) error
}

// RealClock is the wall clock.
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

func (RealClock) Sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// EventKind is what happened to a sink.
type EventKind int

const (
	Dropped  EventKind = iota // the sink was unplugged
	Returned                  // the sink is back
)

// Event reports a sink dropping out or coming back.
type Event struct {
	Sink int
	Kind EventKind
	Err  error // why it dropped, for Dropped
}

const (
	DefaultStep  = 20 * time.Millisecond // matches go-blinkstick's effect step
	DefaultRetry = time.Second
)

// Runner drives an Effect. The zero value is ready to use.
type Runner struct {
	Step     time.Duration // time between frames; 0 means DefaultStep
	Retry    time.Duration // how often to retry an unplugged sink; 0 means DefaultRetry
	Duration time.Duration // stop after this long; 0 means run until ctx ends
	Clock    Clock         // nil means RealClock
	OnEvent  func(Event)   // may be nil
}

// Run renders e onto sinks until ctx ends or Duration passes. It returns
// one error per sink, nil for a sink that ran to the end. A sink failing
// with blinkstick.ErrDisconnected is retried every Retry and never ends the
// run; any other error takes that sink out. Run returns early once every
// sink is out.
func (r Runner) Run(ctx context.Context, e Effect, sinks []Sink) []error {
	step, retry := cmp.Or(r.Step, DefaultStep), cmp.Or(r.Retry, DefaultRetry)
	clock := r.Clock
	if clock == nil {
		clock = RealClock{}
	}
	errs := make([]error, len(sinks))
	down := make([]bool, len(sinks))
	retryAt := make([]time.Time, len(sinks))
	live := len(sinks)
	start := clock.Now()
	for live > 0 && ctx.Err() == nil {
		now := clock.Now()
		t := now.Sub(start)
		if r.Duration > 0 && t >= r.Duration {
			break
		}
		for i, s := range sinks {
			if errs[i] != nil || (down[i] && now.Before(retryAt[i])) {
				continue
			}
			err := s.SetFrame(e.Frame(i, s.LEDs(), t))
			switch {
			case err == nil:
				if down[i] {
					down[i] = false
					r.emit(Event{Sink: i, Kind: Returned})
				}
			case errors.Is(err, blinkstick.ErrDisconnected):
				if !down[i] {
					down[i] = true
					r.emit(Event{Sink: i, Kind: Dropped, Err: err})
				}
				retryAt[i] = now.Add(retry)
			default:
				errs[i] = err
				live--
			}
		}
		if live == 0 || clock.Sleep(ctx, step) != nil {
			break
		}
	}
	return errs
}

func (r Runner) emit(e Event) {
	if r.OnEvent != nil {
		r.OnEvent(e)
	}
}
