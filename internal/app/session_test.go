package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/bawdo/go-blinkstick"

	"github.com/bawdo/blinky/internal/exitcode"
	"github.com/bawdo/blinky/internal/settings"
	"github.com/bawdo/blinky/internal/stick/sticktest"
)

var red = blinkstick.RGB{R: 255}

func setRed(o opened) error { return o.st.SetFrame([]blinkstick.RGB{red, red}) }

func TestEachRunsOnEveryStickAndCloses(t *testing.T) {
	ctl := sticktest.New(sticktest.Nano("BS1", ""), sticktest.Nano("BS2", ""))
	a, _, _ := newTestApp(ctl)
	if err := a.each(infosOf(t, ctl), nil, setRed); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"BS1", "BS2"} {
		if got := ctl.Stick(s).Current(); got[0] != red || !ctl.Stick(s).Closed() {
			t.Errorf("%s: current %v closed %v", s, got, ctl.Stick(s).Closed())
		}
	}
}

func TestEachAppliesSettings(t *testing.T) {
	ctl := sticktest.New(sticktest.Nano("BS1", ""))
	a, _, _ := newTestApp(ctl)
	set := settings.Settings{Brightness: 40, Inverse: true}
	if err := a.each(infosOf(t, ctl), &set, setRed); err != nil {
		t.Fatal(err)
	}
	if s := ctl.Stick("BS1"); s.Limit() != 102 || !s.Inverse() {
		t.Errorf("limit %d inverse %v", s.Limit(), s.Inverse())
	}
}

func TestEachReportsASingleFailureWithItsID(t *testing.T) {
	ctl := sticktest.New(sticktest.Nano("BS1", "desk"))
	ctl.Stick("BS1").FailWith(errors.New("boom"))
	a, _, errb := newTestApp(ctl)
	err := a.each(infosOf(t, ctl), nil, setRed)
	if err == nil || err.Error() != "desk: boom" || errb.Len() != 0 {
		t.Errorf("err %v, stderr %q", err, errb.String())
	}
}

func TestEachReportsPartialFailure(t *testing.T) {
	ctl := sticktest.New(sticktest.Nano("BS1", ""), sticktest.Nano("BS2", ""))
	ctl.Stick("BS2").FailWith(errors.New("boom"))
	a, _, errb := newTestApp(ctl)
	err := a.each(infosOf(t, ctl), nil, setRed)
	if exitcode.From(err) != 6 || err.Error() != "1 of 2 sticks failed" {
		t.Errorf("err %v (exit %d)", err, exitcode.From(err))
	}
	if !strings.Contains(errb.String(), "BS2: boom") {
		t.Errorf("stderr %q", errb.String())
	}
	if ctl.Stick("BS1").Current()[0] != red {
		t.Error("the working stick was not set")
	}
}

func TestEachReportsBusySticks(t *testing.T) {
	ctl := sticktest.New(sticktest.Busy(sticktest.Nano("BS1", "")))
	a, _, _ := newTestApp(ctl)
	err := a.each(infosOf(t, ctl), nil, setRed)
	if exitcode.From(err) != 5 || !strings.Contains(err.Error(), "another program is using it") {
		t.Errorf("err %v (exit %d)", err, exitcode.From(err))
	}
}

func TestEachKeepsTheUnderlyingCodeWhenEveryStickFails(t *testing.T) {
	ctl := sticktest.New(sticktest.Busy(sticktest.Nano("BS1", "")), sticktest.Busy(sticktest.Nano("BS2", "")))
	a, _, _ := newTestApp(ctl)
	err := a.each(infosOf(t, ctl), nil, setRed)
	if exitcode.From(err) != 5 || err.Error() != "all 2 sticks failed" {
		t.Errorf("err %v (exit %d)", err, exitcode.From(err))
	}
}

func TestLibErrorMapsExitCodes(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{nil, 0},
		{blinkstick.ErrNotFound, 4},
		{blinkstick.ErrDisconnected, 4},
		{fmt.Errorf("x: %w", blinkstick.ErrOutOfRange), 2},
		{blinkstick.ErrInvalidName, 2},
		{errors.New("other"), 1},
	}
	for _, tc := range cases {
		if got := exitcode.From(libError(tc.err)); got != tc.want {
			t.Errorf("libError(%v) exits %d, want %d", tc.err, got, tc.want)
		}
	}
}

func TestFinishTurnsOffWhenStopped(t *testing.T) {
	for _, stop := range []error{context.Canceled, context.DeadlineExceeded} {
		ctl := sticktest.New(sticktest.Nano("BS1", ""))
		st, _ := ctl.Open("BS1")
		_ = st.SetFrame([]blinkstick.RGB{red, red})
		if err := finish(st, stop); err != nil {
			t.Errorf("finish(%v) = %v", stop, err)
		}
		if ctl.Stick("BS1").Current()[0] != (blinkstick.RGB{}) {
			t.Errorf("finish(%v) left the LEDs on", stop)
		}
	}
}

func TestFinishPassesOtherErrorsThrough(t *testing.T) {
	ctl := sticktest.New(sticktest.Nano("BS1", ""))
	st, _ := ctl.Open("BS1")
	_ = st.SetFrame([]blinkstick.RGB{red, red})
	boom := errors.New("boom")
	if err := finish(st, boom); !errors.Is(err, boom) {
		t.Errorf("finish(boom) = %v", err)
	}
	if err := finish(st, nil); err != nil {
		t.Errorf("finish(nil) = %v", err)
	}
	if ctl.Stick("BS1").Current()[0] != red {
		t.Error("finish turned off LEDs on a normal end")
	}
}
