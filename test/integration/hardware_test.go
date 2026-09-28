//go:build integration

package integration

import (
	"bytes"
	"context"
	"math/rand/v2"
	"slices"
	"testing"
	"time"

	"github.com/bawdo/go-blinkstick"

	"github.com/bawdo/blinky/internal/app"
	"github.com/bawdo/blinky/internal/colour"
	"github.com/bawdo/blinky/internal/settings"
	"github.com/bawdo/blinky/internal/stick"
	"github.com/bawdo/blinky/internal/target"
)

// These tests write LEDs only, which is RAM and does not wear. They must
// never write EEPROM (names or info blocks).

// attached returns the sticks blinky can drive, or skips the test.
func attached(t *testing.T) []stick.Info {
	t.Helper()
	infos, err := stick.Hardware{}.List()
	if err != nil {
		t.Skipf("cannot list BlinkSticks: %v", err)
	}
	var ok []stick.Info
	for _, i := range infos {
		if i.Status == stick.StatusOK {
			ok = append(ok, i)
		}
	}
	if len(ok) == 0 {
		t.Skip("no BlinkStick attached that blinky can drive")
	}
	return ok
}

func TestHardwareListReportsModels(t *testing.T) {
	for _, i := range attached(t) {
		if (i.Model != "Nano" && i.Model != "Square") || i.LEDs == 0 || i.Serial == "" {
			t.Errorf("unexpected stick %+v", i)
		}
	}
}

func TestHardwareColourRoundTrip(t *testing.T) {
	specs, err := colour.ParseAll([]string{"red", "blue"})
	if err != nil {
		t.Fatal(err)
	}
	r := rand.New(rand.NewPCG(1, 1))
	for _, i := range attached(t) {
		st, err := stick.Hardware{}.Open(i.Serial)
		if err != nil {
			t.Fatalf("%s: %v", i.Serial, err)
		}
		want := colour.Frame(specs, i.LEDs, r)
		if err := st.SetFrame(want); err != nil {
			t.Fatalf("%s: %v", i.Serial, err)
		}
		got, err := st.Frame()
		if err != nil || !slices.Equal(got, want) {
			t.Errorf("%s: read %v, %v; want %v", i.Serial, got, err, want)
		}
		_ = st.Off()
		_ = st.Close()
	}
}

func TestHardwarePoliceRunsAndEndsOff(t *testing.T) {
	infos := attached(t)
	var out, errb bytes.Buffer
	a := app.NewWithOptions(app.Options{Out: &out, Err: &errb})
	err := a.Police(context.Background(), target.Request{All: true}, settings.Default(),
		app.PoliceOptions{Period: 500 * time.Millisecond, Duration: time.Second})
	if err != nil {
		t.Fatalf("police: %v (stderr %q)", err, errb.String())
	}
	for _, i := range infos {
		st, err := stick.Hardware{}.Open(i.Serial)
		if err != nil {
			t.Fatalf("%s: %v", i.Serial, err)
		}
		got, _ := st.Frame()
		_ = st.Close()
		for _, c := range got {
			if c != (blinkstick.RGB{}) {
				t.Errorf("%s left on: %v", i.Serial, got)
				break
			}
		}
	}
}
