package effect

import (
	"slices"
	"testing"
	"time"

	"github.com/bawdo/go-blinkstick"
)

var (
	red    = blinkstick.RGB{R: 255}
	yellow = blinkstick.RGB{R: 255, G: 255}
	dark   = blinkstick.RGB{}
)

// blinkAt checks b's frame for stick 0, two LEDs, at each time.
func blinkAt(t *testing.T, b *Blink, want map[time.Duration]blinkstick.RGB) {
	t.Helper()
	for at, c := range want {
		if got := b.Frame(0, 2, at); !slices.Equal(got, []blinkstick.RGB{c, c}) {
			t.Errorf("t=%v: got %v, want %v", at, got, c)
		}
	}
}

func TestBlinkEvenCycle(t *testing.T) {
	b, err := NewBlink(500*ms, 500*ms, []blinkstick.RGB{red}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if b.Cycle() != time.Second {
		t.Errorf("cycle %v", b.Cycle())
	}
	blinkAt(t, b, map[time.Duration]blinkstick.RGB{0: red, 480 * ms: red, 500 * ms: dark, 980 * ms: dark, time.Second: red})
}

func TestBlinkUnevenCycle(t *testing.T) {
	b, _ := NewBlink(time.Second, 2*time.Second, []blinkstick.RGB{red}, nil)
	blinkAt(t, b, map[time.Duration]blinkstick.RGB{999 * ms: red, time.Second: dark, 2999 * ms: dark, 3 * time.Second: red})
}

func TestBlinkSecondColour(t *testing.T) {
	b, _ := NewBlink(100*ms, 100*ms, []blinkstick.RGB{red}, []blinkstick.RGB{yellow})
	if b.Cycle() != 400*ms {
		t.Errorf("cycle %v", b.Cycle())
	}
	blinkAt(t, b, map[time.Duration]blinkstick.RGB{0: red, 100 * ms: dark, 200 * ms: yellow, 300 * ms: dark, 400 * ms: red})
}

func TestBlinkSecondColourNoGap(t *testing.T) {
	b, _ := NewBlink(100*ms, 0, []blinkstick.RGB{red}, []blinkstick.RGB{yellow})
	blinkAt(t, b, map[time.Duration]blinkstick.RGB{0: red, 100 * ms: yellow, 200 * ms: red})
}

func TestBlinkColourPerStick(t *testing.T) {
	blue := blinkstick.RGB{B: 255}
	b, _ := NewBlink(500*ms, 500*ms, []blinkstick.RGB{red, blue}, nil)
	if got := b.Frame(1, 2, 0); !slices.Equal(got, []blinkstick.RGB{blue, blue}) {
		t.Errorf("stick 1: %v", got)
	}
}

func TestNewBlinkRejects(t *testing.T) {
	one := []blinkstick.RGB{red}
	cases := []struct {
		on, off       time.Duration
		first, second []blinkstick.RGB
	}{
		{0, time.Second, one, nil},
		{time.Second, -1, one, nil},
		{time.Second, time.Second, one, []blinkstick.RGB{red, red}},
	}
	for _, tc := range cases {
		if _, err := NewBlink(tc.on, tc.off, tc.first, tc.second); err == nil {
			t.Errorf("%+v: want an error", tc)
		}
	}
}
