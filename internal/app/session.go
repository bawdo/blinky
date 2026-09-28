package app

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"

	"github.com/bawdo/go-blinkstick"

	"github.com/bawdo/blinky/internal/exitcode"
	"github.com/bawdo/blinky/internal/render"
	"github.com/bawdo/blinky/internal/settings"
	"github.com/bawdo/blinky/internal/stick"
	"github.com/bawdo/blinky/internal/target"
)

// opened is a stick that opened, with its place in the resolved list.
type opened struct {
	pos  int
	info stick.Info
	st   stick.Stick
}

// resolve lists the attached sticks and picks the ones req names. It also
// returns the full list, for checks that look at other sticks.
func (a *App) resolve(req target.Request, scope target.Scope) (chosen, all []stick.Info, err error) {
	all, err = a.ctl.List()
	if err != nil {
		return nil, nil, fmt.Errorf("listing BlinkSticks: %w", err)
	}
	chosen, err = target.Resolve(all, req, scope)
	return chosen, all, err
}

func (a *App) open(info stick.Info) (stick.Stick, error) {
	if info.Status == stick.StatusBusy {
		return nil, fmt.Errorf("%w: another program is using it", exitcode.ErrBusy)
	}
	st, err := a.ctl.Open(info.Serial)
	if err != nil {
		return nil, libError(err)
	}
	return st, nil
}

// session opens every stick in infos, applies set to each if set is not
// nil, runs fn with the ones that opened, closes them and reports every
// failure. fn returns one error per stick it was given, in order.
func (a *App) session(infos []stick.Info, set *settings.Settings, fn func([]opened) []error) error {
	errs := make([]error, len(infos))
	var sticks []opened
	for i, info := range infos {
		st, err := a.open(info)
		if err != nil {
			errs[i] = err
			continue
		}
		if set != nil {
			set.Apply(st)
		}
		sticks = append(sticks, opened{pos: i, info: info, st: st})
	}
	defer func() {
		for _, o := range sticks {
			_ = o.st.Close()
		}
	}()
	if len(sticks) > 0 {
		for j, err := range fn(sticks) {
			errs[sticks[j].pos] = err
		}
	}
	return a.report(infos, errs)
}

// each runs fn on every opened stick at once.
func (a *App) each(infos []stick.Info, set *settings.Settings, fn func(opened) error) error {
	return a.session(infos, set, func(sticks []opened) []error {
		errs := make([]error, len(sticks))
		var wg sync.WaitGroup
		for j, o := range sticks {
			wg.Go(func() { errs[j] = fn(o) })
		}
		wg.Wait()
		return errs
	})
}

// report turns per-stick errors into the command's result. With one stick
// it is that stick's error, prefixed with its ID. With several, each
// failure goes to stderr and the result is ErrPartial if any stick worked.
func (a *App) report(infos []stick.Info, errs []error) error {
	var failed []error
	for i, err := range errs {
		if err != nil {
			failed = append(failed, fmt.Errorf("%s: %w", render.Sanitise(infos[i].ID()), err))
		}
	}
	switch {
	case len(failed) == 0:
		return nil
	case len(infos) == 1:
		return failed[0]
	}
	for _, err := range failed {
		_, _ = fmt.Fprintln(a.err, err)
	}
	if len(failed) < len(infos) {
		return &groupError{
			msg:  fmt.Sprintf("%d of %d sticks failed", len(failed), len(infos)),
			errs: append([]error{exitcode.ErrPartial}, failed...),
		}
	}
	return &groupError{msg: fmt.Sprintf("all %d sticks failed", len(infos)), errs: failed}
}

// groupError summarises several stick failures while errors.Is still sees
// every one of them.
type groupError struct {
	msg  string
	errs []error
}

func (e *groupError) Error() string   { return e.msg }
func (e *groupError) Unwrap() []error { return e.errs }

// printRows prints read results: JSON as one array, or one "ID: value"
// line per stick.
func printRows[T any](a *App, asJSON bool, rows []T, line func(T) (id, value string)) error {
	if asJSON {
		return render.JSON(a.out, rows)
	}
	for _, r := range rows {
		id, value := line(r)
		if _, err := fmt.Fprintf(a.out, "%s: %s\n", render.Sanitise(id), value); err != nil {
			return err
		}
	}
	return nil
}

// libError maps go-blinkstick errors onto exit code sentinels.
func libError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, blinkstick.ErrNotFound): // also matches ErrDisconnected
		return fmt.Errorf("%w: %w", exitcode.ErrNotFound, err)
	case errors.Is(err, blinkstick.ErrOutOfRange), errors.Is(err, blinkstick.ErrInvalidName):
		return fmt.Errorf("%w: %w", exitcode.ErrInvalidArgs, err)
	}
	return err
}

// stopped reports whether err is the context ending: Ctrl-C or --duration.
func stopped(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// finish ends an animated command on one stick. A stop by Ctrl-C or
// --duration turns the LEDs off and counts as success.
func finish(st stick.Stick, err error) error {
	if err == nil {
		return nil
	}
	if !stopped(err) {
		return libError(err)
	}
	return libError(st.Off())
}

// newRand returns a generator for colour and effect choices. Seed 0 picks
// a random seed.
func newRand(seed uint64) *rand.Rand {
	if seed == 0 {
		seed = rand.Uint64() //nolint:gosec // G404: LED colours, not security
	}
	return rand.New(rand.NewPCG(seed, seed)) //nolint:gosec // G404: LED colours, not security
}
