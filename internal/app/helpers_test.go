package app

import (
	"bytes"
	"testing"

	"github.com/bawdo/blinky/internal/effect/effecttest"
	"github.com/bawdo/blinky/internal/stick"
)

// newTestApp returns an App on ctl with captured output and a manual
// clock.
func newTestApp(ctl stick.Controller) (*App, *bytes.Buffer, *bytes.Buffer) {
	var out, errb bytes.Buffer
	a := NewWithOptions(Options{Out: &out, Err: &errb, Controller: ctl, Clock: effecttest.NewClock()})
	return a, &out, &errb
}

// infosOf lists ctl's sticks or fails the test.
func infosOf(t *testing.T, ctl stick.Controller) []stick.Info {
	t.Helper()
	infos, err := ctl.List()
	if err != nil {
		t.Fatal(err)
	}
	return infos
}
