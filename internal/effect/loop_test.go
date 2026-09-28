package effect

import (
	"slices"
	"testing"
	"time"

	"github.com/bawdo/go-blinkstick"
)

func TestLoopRoundTrip(t *testing.T) {
	blue := blinkstick.RGB{B: 255}
	l, err := NewLoop(time.Second, [][]blinkstick.RGB{{red}}, []blinkstick.RGB{blue})
	if err != nil {
		t.Fatal(err)
	}
	if l.Cycle() != 2*time.Second {
		t.Errorf("cycle %v", l.Cycle())
	}
	cases := []struct {
		t    time.Duration
		want blinkstick.RGB
	}{
		{0, red},
		{500 * ms, blinkstick.RGB{R: 128, B: 127}},
		{time.Second, blue},
		{1500 * ms, blinkstick.RGB{R: 127, B: 128}},
		{2 * time.Second, red},
	}
	for _, tc := range cases {
		if got := l.Frame(0, 1, tc.t); !slices.Equal(got, []blinkstick.RGB{tc.want}) {
			t.Errorf("t=%v: got %v, want %v", tc.t, got, tc.want)
		}
	}
}

func TestLoopKeepsMixedStartingLEDs(t *testing.T) {
	green := blinkstick.RGB{G: 255}
	l, _ := NewLoop(time.Second, [][]blinkstick.RGB{{red, green}}, []blinkstick.RGB{dark})
	if got := l.Frame(0, 2, 0); !slices.Equal(got, []blinkstick.RGB{red, green}) {
		t.Errorf("start: %v", got)
	}
	if got := l.Frame(0, 4, 2*time.Second); !slices.Equal(got, []blinkstick.RGB{red, green, red, green}) {
		t.Errorf("repeated to fit 4 LEDs: %v", got)
	}
}

func TestLoopEmptyFromStartsDark(t *testing.T) {
	l, _ := NewLoop(time.Second, [][]blinkstick.RGB{nil}, []blinkstick.RGB{red})
	if got := l.Frame(0, 2, 0); !slices.Equal(got, []blinkstick.RGB{dark, dark}) {
		t.Errorf("got %v", got)
	}
}

func TestNewLoopRejects(t *testing.T) {
	if _, err := NewLoop(0, [][]blinkstick.RGB{{red}}, []blinkstick.RGB{red}); err == nil {
		t.Error("want an error for a zero fade")
	}
	if _, err := NewLoop(time.Second, [][]blinkstick.RGB{{red}}, []blinkstick.RGB{red, red}); err == nil {
		t.Error("want an error for mismatched sticks")
	}
}
