// Package app is the orchestrator behind every CLI command.
//
// Convention: cmd/ files are thin (Cobra wiring, flag parsing,
// delegation). All business logic - validation, I/O, state mutation -
// lives here so it can be unit-tested without spinning up Cobra. The App
// is built once, in newRootCmd, and each command closes over it. A cmd
// function should look like:
//
//	func newThingCmd(a *app.App) *cobra.Command {
//	    c := &cobra.Command{ /* ... */ }
//	    c.RunE = func(cmd *cobra.Command, args []string) error {
//	        return a.DoTheThing(cmd.Context(), parseInput(args))
//	    }
//	    return c
//	}
//
// Errors returned from App methods that wrap an exitcode sentinel
// (see internal/exitcode) will translate to the right exit code on
// the way out.
package app

import (
	"io"
	"os"

	"github.com/bawdo/blinky/internal/effect"
	"github.com/bawdo/blinky/internal/stick"
)

// App owns the runtime dependencies the commands share.
type App struct {
	out   io.Writer
	err   io.Writer
	ctl   stick.Controller
	clock effect.Clock
}

// Options is the test-friendly constructor input. Zero-valued fields
// fall back to production defaults; tests fill in only the bits they
// care about.
type Options struct {
	// Out is where command results go. Defaults to os.Stdout.
	Out io.Writer
	// Err is where warnings and per-stick errors go. Defaults to os.Stderr.
	Err io.Writer
	// Controller finds and opens sticks. Defaults to stick.Hardware{}.
	Controller stick.Controller
	// Clock drives disco and police. Defaults to effect.RealClock{}.
	Clock effect.Clock
}

// New builds an App with default dependencies. Production callers use
// this; tests use NewWithOptions.
func New() *App {
	return NewWithOptions(Options{})
}

// NewWithOptions builds an App from explicit options, substituting
// production defaults for any zero-valued field.
func NewWithOptions(opts Options) *App {
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	if opts.Err == nil {
		opts.Err = os.Stderr
	}
	if opts.Controller == nil {
		opts.Controller = stick.Hardware{}
	}
	if opts.Clock == nil {
		opts.Clock = effect.RealClock{}
	}
	return &App{out: opts.Out, err: opts.Err, ctl: opts.Controller, clock: opts.Clock}
}

// Out returns the writer App was constructed with. cmd/ render helpers
// use this so test output goes to the same buffer the test injected.
func (a *App) Out() io.Writer { return a.out }

// Err returns the writer for warnings and per-stick errors.
func (a *App) Err() io.Writer { return a.err }
