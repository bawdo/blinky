package app

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bawdo/go-blinkstick"

	"github.com/bawdo/blinky/internal/colour"
	"github.com/bawdo/blinky/internal/exitcode"
)

const ms = time.Millisecond

func TestPoliceRunsForDurationThenTurnsOff(t *testing.T) {
	ctl := deskAndSquare()
	a, _, _ := newTestApp(ctl)
	if err := a.Police(context.Background(), all, full, PoliceOptions{Period: time.Second, Duration: 100 * ms}); err != nil {
		t.Fatal(err)
	}
	frames := ctl.Stick("BS072777-3.0").Frames()
	if len(frames) != 6 { // 0, 20, 40, 60, 80ms, then off
		t.Fatalf("got %d frames", len(frames))
	}
	if frames[0][1] != (blinkstick.RGB{B: 255}) || frames[5][1] != (blinkstick.RGB{}) {
		t.Errorf("first %v last %v", frames[0], frames[5])
	}
}

func TestPoliceAlternate(t *testing.T) {
	ctl := deskAndSquare()
	a, _, _ := newTestApp(ctl)
	_ = a.Police(context.Background(), all, full, PoliceOptions{Period: time.Second, Duration: 20 * ms, Alternate: true})
	first := ctl.Stick("BS073788-3.1").Frames()[0]
	if first[0] != (blinkstick.RGB{B: 255}) || first[7] != (blinkstick.RGB{}) {
		t.Errorf("second stick not swapped: %v", first)
	}
}

func TestDiscoUsesThePalette(t *testing.T) {
	ctl := deskAndSquare()
	a, _, _ := newTestApp(ctl)
	o := DiscoOptions{MinPeriod: 200 * ms, MaxPeriod: 2 * time.Second, MaxGap: time.Second, Duration: 5 * time.Second,
		Palette: []colour.Spec{{RGB: blinkstick.RGB{R: 255}}}, Seed: 1}
	if err := a.Disco(context.Background(), all, full, o); err != nil {
		t.Fatal(err)
	}
	lit := false
	for _, f := range ctl.Stick("BS073788-3.1").Frames() {
		for _, c := range f {
			if c.G != 0 || c.B != 0 {
				t.Fatalf("colour outside the palette: %v", c)
			}
			lit = lit || c.R > 0
		}
	}
	if !lit {
		t.Error("nothing lit up in 5 seconds")
	}
}

func TestDiscoRepeatsWithASeed(t *testing.T) {
	runOnce := func() [][]blinkstick.RGB {
		ctl := deskAndSquare()
		a, _, _ := newTestApp(ctl)
		o := DiscoOptions{MinPeriod: 200 * ms, MaxPeriod: 2 * time.Second, MaxGap: time.Second, Duration: 2 * time.Second, Seed: 7}
		if err := a.Disco(context.Background(), all, full, o); err != nil {
			t.Fatal(err)
		}
		return ctl.Stick("BS073788-3.1").Frames()
	}
	if a, b := runOnce(), runOnce(); !reflect.DeepEqual(a, b) {
		t.Error("same seed gave different frames")
	}
}

func TestModesValidate(t *testing.T) {
	a, _, _ := newTestApp(deskAndSquare())
	ctx := context.Background()
	if err := a.Disco(ctx, all, full, DiscoOptions{MaxPeriod: time.Second}); exitcode.From(err) != 2 {
		t.Errorf("disco: err %v", err)
	}
	if err := a.Police(ctx, all, full, PoliceOptions{}); exitcode.From(err) != 2 {
		t.Errorf("police: err %v", err)
	}
	if err := a.Police(ctx, all, full, PoliceOptions{Period: time.Second, Duration: -1}); exitcode.From(err) != 2 {
		t.Errorf("negative duration: err %v", err)
	}
}

func TestModesReportDisconnectedSticks(t *testing.T) {
	ctl := deskAndSquare()
	ctl.Stick("BS072777-3.0").FailWith(blinkstick.ErrDisconnected)
	a, _, errb := newTestApp(ctl)
	err := a.Police(context.Background(), all, full, PoliceOptions{Period: time.Second, Duration: 100 * ms})
	if exitcode.From(err) != 6 || !strings.Contains(errb.String(), "desk: disconnected, retrying") {
		t.Errorf("err %v stderr %q", err, errb.String())
	}
}
