package app

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/bawdo/go-blinkstick"

	"github.com/bawdo/blinky/internal/colour"
	"github.com/bawdo/blinky/internal/exitcode"
	"github.com/bawdo/blinky/internal/settings"
	"github.com/bawdo/blinky/internal/stick/sticktest"
	"github.com/bawdo/blinky/internal/target"
)

var (
	blue = blinkstick.RGB{B: 255}
	all  = target.Request{All: true}
	full = settings.Default()
)

func deskAndSquare() *sticktest.Controller {
	return sticktest.New(sticktest.Nano("BS072777-3.0", "desk"), sticktest.Square("BS073788-3.1", ""))
}

func specs(t *testing.T, args ...string) []colour.Spec {
	t.Helper()
	s, err := colour.ParseAll(args)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestSetColourMapsAcrossSticks(t *testing.T) {
	ctl := deskAndSquare()
	a, _, _ := newTestApp(ctl)
	if err := a.SetColour(all, full, -1, specs(t, "red", "blue")); err != nil {
		t.Fatal(err)
	}
	if got := ctl.Stick("BS072777-3.0").Current(); !slices.Equal(got, []blinkstick.RGB{red, blue}) {
		t.Errorf("Nano %v", got)
	}
	want := []blinkstick.RGB{red, red, red, red, blue, blue, blue, blue}
	if got := ctl.Stick("BS073788-3.1").Current(); !slices.Equal(got, want) {
		t.Errorf("Square %v", got)
	}
}

func TestSetColourOneLED(t *testing.T) {
	ctl := deskAndSquare()
	a, _, _ := newTestApp(ctl)
	req := target.Request{Devices: []string{"BS073788-3.1"}}
	if err := a.SetColour(req, full, 5, specs(t, "red")); err != nil {
		t.Fatal(err)
	}
	got := ctl.Stick("BS073788-3.1").Current()
	if got[5] != red || got[4] != (blinkstick.RGB{}) {
		t.Errorf("Square %v", got)
	}
}

func TestSetColourLEDMustExistOnEveryStick(t *testing.T) {
	a, _, _ := newTestApp(deskAndSquare())
	err := a.SetColour(all, full, 5, specs(t, "red"))
	if exitcode.From(err) != 2 || !strings.Contains(err.Error(), "LED 5 does not exist on desk") {
		t.Errorf("err %v", err)
	}
}

func TestSetColourLEDTakesOneColour(t *testing.T) {
	a, _, _ := newTestApp(deskAndSquare())
	if err := a.SetColour(all, full, 0, specs(t, "red", "blue")); exitcode.From(err) != 2 {
		t.Errorf("err %v", err)
	}
}

func TestSetColourAppliesBrightness(t *testing.T) {
	ctl := deskAndSquare()
	a, _, _ := newTestApp(ctl)
	if err := a.SetColour(all, settings.Settings{Brightness: 50}, -1, specs(t, "red")); err != nil {
		t.Fatal(err)
	}
	if got := ctl.Stick("BS072777-3.0").Current()[0]; got != (blinkstick.RGB{R: 128}) {
		t.Errorf("got %v", got)
	}
}

func TestSetColourVividPicksPerGroup(t *testing.T) {
	ctl := deskAndSquare()
	a, _, _ := newTestApp(ctl)
	if err := a.SetColour(all, full, -1, specs(t, "vivid", "vivid")); err != nil {
		t.Fatal(err)
	}
	sq := ctl.Stick("BS073788-3.1").Current()
	for i := 1; i < 4; i++ {
		if sq[i] != sq[0] || sq[i+4] != sq[4] {
			t.Fatalf("groups not uniform: %v", sq)
		}
	}
	if max(sq[0].R, sq[0].G, sq[0].B) != 255 {
		t.Errorf("not vivid: %v", sq[0])
	}
}

func TestReadColour(t *testing.T) {
	ctl := deskAndSquare()
	a, out, _ := newTestApp(ctl)
	_ = a.SetColour(all, full, -1, specs(t, "red", "blue"))
	if err := a.ReadColour(all, full, -1, false); err != nil {
		t.Fatal(err)
	}
	want := "desk: #ff0000 #0000ff\n" +
		"BS073788-3.1: #ff0000 #ff0000 #ff0000 #ff0000 #0000ff #0000ff #0000ff #0000ff\n"
	if out.String() != want {
		t.Errorf("got %q", out.String())
	}
}

func TestReadColourOneLEDAsJSON(t *testing.T) {
	ctl := deskAndSquare()
	a, out, _ := newTestApp(ctl)
	_ = a.SetColour(all, full, -1, specs(t, "red", "blue"))
	if err := a.ReadColour(all, full, 1, true); err != nil {
		t.Fatal(err)
	}
	var got []struct {
		ID      string   `json:"id"`
		Serial  string   `json:"serial"`
		Colours []string `json:"colours"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != "desk" || !slices.Equal(got[0].Colours, []string{"#0000ff"}) ||
		!slices.Equal(got[1].Colours, []string{"#ff0000"}) {
		t.Errorf("got %+v", got)
	}
}

func TestOff(t *testing.T) {
	ctl := deskAndSquare()
	a, _, _ := newTestApp(ctl)
	_ = a.SetColour(all, full, -1, specs(t, "red"))
	if err := a.Off(all, full); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"BS072777-3.0", "BS073788-3.1"} {
		for _, c := range ctl.Stick(s).Current() {
			if c != (blinkstick.RGB{}) {
				t.Errorf("%s still lit: %v", s, ctl.Stick(s).Current())
			}
		}
	}
}
