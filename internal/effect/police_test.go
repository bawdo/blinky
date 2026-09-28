package effect

import (
	"slices"
	"testing"
	"time"

	"github.com/bawdo/go-blinkstick"
)

func TestPoliceNanoCycle(t *testing.T) {
	p, err := NewPolice(time.Second, false)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		t    time.Duration
		want []blinkstick.RGB
	}{
		{0, []blinkstick.RGB{{}, {B: 255}}},
		{250 * ms, []blinkstick.RGB{{R: 127}, {B: 128}}},
		{500 * ms, []blinkstick.RGB{{R: 255}, {}}},
		{750 * ms, []blinkstick.RGB{{R: 128}, {B: 127}}},
		{time.Second, []blinkstick.RGB{{}, {B: 255}}},
	}
	for _, tc := range cases {
		if got := p.Frame(0, 2, tc.t); !slices.Equal(got, tc.want) {
			t.Errorf("t=%v: got %v, want %v", tc.t, got, tc.want)
		}
	}
}

func TestPoliceSquareHalves(t *testing.T) {
	p, _ := NewPolice(time.Second, false)
	got := p.Frame(0, 8, 500*ms)
	for i, c := range got {
		want := blinkstick.RGB{}
		if i < 4 {
			want = blinkstick.RGB{R: 255}
		}
		if c != want {
			t.Errorf("LED %d = %v, want %v", i, c, want)
		}
	}
}

func TestPoliceAlternateSwapsOddSticks(t *testing.T) {
	p, _ := NewPolice(time.Second, true)
	if got := p.Frame(0, 2, 0); !slices.Equal(got, []blinkstick.RGB{{}, {B: 255}}) {
		t.Errorf("stick 0: %v", got)
	}
	if got := p.Frame(1, 2, 0); !slices.Equal(got, []blinkstick.RGB{{B: 255}, {}}) {
		t.Errorf("stick 1: %v", got)
	}
	if got := p.Frame(2, 2, 0); !slices.Equal(got, []blinkstick.RGB{{}, {B: 255}}) {
		t.Errorf("stick 2: %v", got)
	}
}

func TestNewPoliceRejectsZeroPeriod(t *testing.T) {
	if _, err := NewPolice(0, false); err == nil {
		t.Error("want an error for a zero period")
	}
}
