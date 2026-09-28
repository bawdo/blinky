package app

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/bawdo/go-blinkstick"

	"github.com/bawdo/blinky/internal/colour"
	"github.com/bawdo/blinky/internal/exitcode"
	"github.com/bawdo/blinky/internal/settings"
	"github.com/bawdo/blinky/internal/stick/sticktest"
)

func oneNano() *sticktest.Controller { return sticktest.New(sticktest.Nano("BS1", "desk")) }

func redSpec() colour.Spec { return colour.Spec{RGB: blinkstick.RGB{R: 255}} }

var (
	yellow = blinkstick.RGB{R: 255, G: 255}
	dark   = blinkstick.RGB{}
)

func dur(d time.Duration) *time.Duration { return &d }

// checkFrames compares LED 0 of chosen frames.
func checkFrames(t *testing.T, frames [][]blinkstick.RGB, want map[int]blinkstick.RGB) {
	t.Helper()
	for i, c := range want {
		if frames[i][0] != c {
			t.Errorf("frame %d: got %v, want %v", i, frames[i][0], c)
		}
	}
}

func TestBlinkDefaultCycle(t *testing.T) {
	ctl := oneNano()
	a, _, _ := newTestApp(ctl)
	if err := a.Blink(context.Background(), all, full, redSpec(), BlinkOptions{Period: time.Second, Repeats: 3}); err != nil {
		t.Fatal(err)
	}
	frames := ctl.Stick("BS1").Frames()
	if len(frames) != 151 { // every 20ms for 3s, then off
		t.Fatalf("got %d frames", len(frames))
	}
	checkFrames(t, frames, map[int]blinkstick.RGB{0: red, 24: red, 25: dark, 49: dark, 50: red, 100: red, 149: dark, 150: dark})
}

func TestBlinkOnOffAndSecondColour(t *testing.T) {
	ctl := oneNano()
	a, _, _ := newTestApp(ctl)
	second := colour.Spec{RGB: yellow}
	o := BlinkOptions{Period: time.Second, On: dur(100 * ms), Off: dur(200 * ms), Second: &second, Repeats: 1}
	if err := a.Blink(context.Background(), all, full, redSpec(), o); err != nil {
		t.Fatal(err)
	}
	frames := ctl.Stick("BS1").Frames()
	if len(frames) != 31 { // one 600ms cycle, then off
		t.Fatalf("got %d frames", len(frames))
	}
	checkFrames(t, frames, map[int]blinkstick.RGB{0: red, 4: red, 5: dark, 14: dark, 15: yellow, 19: yellow, 20: dark, 30: dark})
}

func TestBlinkOnTimeKeepsHalfPeriodOff(t *testing.T) {
	ctl := oneNano()
	a, _, _ := newTestApp(ctl)
	o := BlinkOptions{Period: 2 * time.Second, On: dur(300 * ms), Repeats: 1}
	if err := a.Blink(context.Background(), all, full, redSpec(), o); err != nil {
		t.Fatal(err)
	}
	if n := len(ctl.Stick("BS1").Frames()); n != 66 { // 300ms on + 1s off, then off
		t.Errorf("got %d frames", n)
	}
}

func TestBlinkDurationCapsRepeats(t *testing.T) {
	ctl := oneNano()
	a, _, _ := newTestApp(ctl)
	o := BlinkOptions{Period: time.Second, Repeats: 3, Duration: 100 * ms}
	if err := a.Blink(context.Background(), all, full, redSpec(), o); err != nil {
		t.Fatal(err)
	}
	s := ctl.Stick("BS1")
	if n := len(s.Frames()); n != 6 {
		t.Errorf("got %d frames", n)
	}
	if s.Current()[0] != dark {
		t.Errorf("left on: %v", s.Current())
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
	if err := a.Morph(context.Background(), all, full, redSpec(), MorphOptions{Fade: time.Second}); err != nil {
		t.Fatal(err)
	}
	s := ctl.Stick("BS1")
	if !slices.Equal(s.Calls(), []string{"morph #ff0000 1s"}) || s.Current()[0] != red {
		t.Errorf("calls %v current %v", s.Calls(), s.Current())
	}
}

func TestMorphFromColourSetsThenFades(t *testing.T) {
	ctl := oneNano()
	a, _, _ := newTestApp(ctl)
	from := colour.Spec{RGB: yellow}
	if err := a.Morph(context.Background(), all, full, redSpec(), MorphOptions{From: &from, Fade: time.Second}); err != nil {
		t.Fatal(err)
	}
	s := ctl.Stick("BS1")
	if f := s.Frames(); len(f) == 0 || !slices.Equal(f[0], []blinkstick.RGB{yellow, yellow}) {
		t.Errorf("frames %v", f)
	}
	if !slices.Equal(s.Calls(), []string{"morph #ff0000 1s"}) || s.Current()[0] != red {
		t.Errorf("calls %v current %v", s.Calls(), s.Current())
	}
}

func TestMorphDurationCutsFadeShortAndTurnsOff(t *testing.T) {
	ctl := oneNano()
	s := ctl.Stick("BS1")
	s.BlockEffects()
	a, _, _ := newTestApp(ctl)
	if err := a.Morph(context.Background(), all, full, redSpec(), MorphOptions{Fade: time.Second, Duration: 30 * ms}); err != nil {
		t.Fatalf("err %v", err)
	}
	if s.Current()[0] != dark {
		t.Errorf("left on: %v", s.Current())
	}
}

func TestMorphLoopFromWhatTheLEDsShow(t *testing.T) {
	ctl := oneNano()
	s := ctl.Stick("BS1")
	green := blinkstick.RGB{G: 255} // blue is declared in colour_test.go
	_ = s.SetFrame([]blinkstick.RGB{green, blue})
	a, _, _ := newTestApp(ctl)
	if err := a.Morph(context.Background(), all, full, redSpec(), MorphOptions{Fade: 100 * ms, Loop: true, Repeats: 2}); err != nil {
		t.Fatal(err)
	}
	frames := s.Frames() // the frame set above, 20 at 20ms over 400ms, then off
	if len(frames) != 22 {
		t.Fatalf("got %d frames", len(frames))
	}
	for i, want := range map[int][]blinkstick.RGB{
		1:  {green, blue},
		6:  {red, red},
		11: {green, blue},
		21: {dark, dark},
	} {
		if !slices.Equal(frames[i], want) {
			t.Errorf("frame %d: got %v, want %v", i, frames[i], want)
		}
	}
}

func TestMorphLoopFromColour(t *testing.T) {
	ctl := oneNano()
	a, _, _ := newTestApp(ctl)
	from := colour.Spec{RGB: yellow}
	if err := a.Morph(context.Background(), all, full, redSpec(), MorphOptions{From: &from, Fade: 100 * ms, Loop: true, Repeats: 1}); err != nil {
		t.Fatal(err)
	}
	frames := ctl.Stick("BS1").Frames()
	checkFrames(t, frames, map[int]blinkstick.RGB{0: yellow, 5: red, 10: dark})
}

func TestMorphLoopDoesNotDimTwice(t *testing.T) {
	ctl := oneNano()
	s := ctl.Stick("BS1")
	s.SetBrightnessLimit(128) // a previous run at 50 per cent left it showing R 128
	_ = s.SetFrame([]blinkstick.RGB{red, red})
	half := settings.Settings{Brightness: 50}
	a, _, _ := newTestApp(ctl)
	if err := a.Morph(context.Background(), all, half, redSpec(), MorphOptions{Fade: 100 * ms, Loop: true, Repeats: 1}); err != nil {
		t.Fatal(err)
	}
	if got := s.Frames()[1][0]; got != (blinkstick.RGB{R: 128}) {
		t.Errorf("start frame %v, want R 128", got)
	}
}

func TestMorphLoopDurationCapsRepeats(t *testing.T) {
	ctl := oneNano()
	a, _, _ := newTestApp(ctl)
	if err := a.Morph(context.Background(), all, full, redSpec(), MorphOptions{Fade: 100 * ms, Loop: true, Duration: 60 * ms}); err != nil {
		t.Fatal(err)
	}
	if n := len(ctl.Stick("BS1").Frames()); n != 4 {
		t.Errorf("got %d frames", n)
	}
}

func TestMorphValidates(t *testing.T) {
	a, _, _ := newTestApp(oneNano())
	bad := []MorphOptions{
		{Fade: -1},
		{Fade: time.Second, Loop: true, Repeats: -1},
		{Fade: time.Second, Duration: -1},
		{Fade: time.Second, Repeats: 2},
		{Loop: true},
	}
	for _, m := range bad {
		if err := a.Morph(context.Background(), all, full, redSpec(), m); exitcode.From(err) != 2 {
			t.Errorf("%+v: err %v", m, err)
		}
	}
}

func TestBlinkForeverStopsAtDurationAndTurnsOff(t *testing.T) {
	ctl := oneNano()
	s := ctl.Stick("BS1")
	a, _, _ := newTestApp(ctl)
	if err := a.Blink(context.Background(), all, full, redSpec(), BlinkOptions{Period: time.Second, Duration: 30 * ms}); err != nil {
		t.Fatalf("err %v", err)
	}
	if s.Current()[0] != dark {
		t.Errorf("left on: %v", s.Current())
	}
}

func TestBlinkValidates(t *testing.T) {
	a, _, _ := newTestApp(oneNano())
	bad := []BlinkOptions{
		{Period: 0, Repeats: 1},
		{Period: time.Second, Repeats: -1},
		{Period: time.Second, Duration: -1},
		{Period: time.Second, On: dur(0)},
		{Period: time.Second, Off: dur(-1)},
	}
	for _, o := range bad {
		if err := a.Blink(context.Background(), all, full, redSpec(), o); exitcode.From(err) != 2 {
			t.Errorf("%+v: err %v", o, err)
		}
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
		if err := a.Pulse(context.Background(), all, full, redSpec(), r); exitcode.From(err) != 2 {
			t.Errorf("%+v: err %v", r, err)
		}
	}
}
