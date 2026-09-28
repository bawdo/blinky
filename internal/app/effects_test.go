package app

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/bawdo/go-blinkstick"

	"github.com/bawdo/blinky/internal/colour"
	"github.com/bawdo/blinky/internal/exitcode"
	"github.com/bawdo/blinky/internal/stick/sticktest"
)

func oneNano() *sticktest.Controller { return sticktest.New(sticktest.Nano("BS1", "desk")) }

func redSpec() colour.Spec { return colour.Spec{RGB: blinkstick.RGB{R: 255}} }

func TestBlinkRepeats(t *testing.T) {
	ctl := oneNano()
	a, _, _ := newTestApp(ctl)
	if err := a.Blink(context.Background(), all, full, redSpec(), Repeat{Period: time.Second, Repeats: 3}); err != nil {
		t.Fatal(err)
	}
	want := []string{"blink #ff0000 1s 1", "blink #ff0000 1s 1", "blink #ff0000 1s 1"}
	if got := ctl.Stick("BS1").Calls(); !slices.Equal(got, want) {
		t.Errorf("calls %v", got)
	}
}

func TestPulseUsesPeriod(t *testing.T) {
	ctl := oneNano()
	a, _, _ := newTestApp(ctl)
	if err := a.Pulse(context.Background(), all, full, redSpec(), Repeat{Period: 2 * time.Second, Repeats: 1}); err != nil {
		t.Fatal(err)
	}
	if got := ctl.Stick("BS1").Calls(); !slices.Equal(got, []string{"pulse #ff0000 2s 1"}) {
		t.Errorf("calls %v", got)
	}
}

func TestMorphEndsOnItsColour(t *testing.T) {
	ctl := oneNano()
	a, _, _ := newTestApp(ctl)
	if err := a.Morph(context.Background(), all, full, redSpec(), time.Second); err != nil {
		t.Fatal(err)
	}
	s := ctl.Stick("BS1")
	if !slices.Equal(s.Calls(), []string{"morph #ff0000 1s"}) || s.Current()[0] != (blinkstick.RGB{R: 255}) {
		t.Errorf("calls %v current %v", s.Calls(), s.Current())
	}
}

func TestBlinkForeverStopsAtDurationAndTurnsOff(t *testing.T) {
	ctl := oneNano()
	s := ctl.Stick("BS1")
	_ = s.SetFrame([]blinkstick.RGB{{R: 255}, {R: 255}})
	s.BlockEffects()
	a, _, _ := newTestApp(ctl)
	err := a.Blink(context.Background(), all, full, redSpec(), Repeat{Period: time.Second, Duration: 30 * time.Millisecond})
	if err != nil {
		t.Fatalf("err %v", err)
	}
	if s.Current()[0] != (blinkstick.RGB{}) {
		t.Errorf("left on: %v", s.Current())
	}
}

func TestPulseStopsOnCancel(t *testing.T) {
	ctl := oneNano()
	ctl.Stick("BS1").BlockEffects()
	a, _, _ := newTestApp(ctl)
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(10*time.Millisecond, cancel)
	if err := a.Pulse(ctx, all, full, redSpec(), Repeat{Period: time.Second}); err != nil {
		t.Errorf("err %v", err)
	}
}

func TestRepeatValidates(t *testing.T) {
	a, _, _ := newTestApp(oneNano())
	bad := []Repeat{{Period: 0, Repeats: 1}, {Period: time.Second, Repeats: -1}, {Period: time.Second, Duration: -1}}
	for _, r := range bad {
		if err := a.Blink(context.Background(), all, full, redSpec(), r); exitcode.From(err) != 2 {
			t.Errorf("%+v: err %v", r, err)
		}
	}
	if err := a.Morph(context.Background(), all, full, redSpec(), -1); exitcode.From(err) != 2 {
		t.Errorf("morph: err %v", err)
	}
}
