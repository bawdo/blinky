package effect

import (
	"math/rand/v2"
	"slices"
	"testing"
	"time"

	"github.com/bawdo/go-blinkstick"
)

func seeded(seed uint64) *rand.Rand { return rand.New(rand.NewPCG(seed, seed)) }

func white(*rand.Rand) blinkstick.RGB { return blinkstick.White }

func TestNewDiscoValidates(t *testing.T) {
	ok := DiscoConfig{MinPeriod: 200 * ms, MaxPeriod: 2 * time.Second, MaxGap: time.Second, Colour: white, Rand: seeded(1)}
	if _, err := NewDisco(ok); err != nil {
		t.Fatalf("valid config: %v", err)
	}
	bad := []func(c *DiscoConfig){
		func(c *DiscoConfig) { c.MinPeriod = 0 },
		func(c *DiscoConfig) { c.MaxPeriod = 100 * ms },
		func(c *DiscoConfig) { c.MaxGap = -1 },
		func(c *DiscoConfig) { c.Colour = nil },
		func(c *DiscoConfig) { c.Rand = nil },
	}
	for i, f := range bad {
		c := ok
		f(&c)
		if _, err := NewDisco(c); err == nil {
			t.Errorf("case %d: want an error", i)
		}
	}
}

func TestDiscoPulseShape(t *testing.T) {
	d, _ := NewDisco(DiscoConfig{MinPeriod: 100 * ms, MaxPeriod: 100 * ms, Colour: white, Rand: seeded(1)})
	cases := map[time.Duration]uint8{0: 0, 25 * ms: 127, 50 * ms: 255, 75 * ms: 127, 100 * ms: 0}
	for _, at := range []time.Duration{0, 25 * ms, 50 * ms, 75 * ms, 100 * ms} {
		want := cases[at]
		if got := d.Frame(0, 1, at)[0]; got != (blinkstick.RGB{R: want, G: want, B: want}) {
			t.Errorf("t=%v: got %v, want level %d", at, got, want)
		}
	}
}

func TestDiscoDarkBeforeStartAndDuringGap(t *testing.T) {
	d, _ := NewDisco(DiscoConfig{MinPeriod: 100 * ms, MaxPeriod: 100 * ms, MaxGap: time.Second, Colour: white, Rand: seeded(2)})
	d.Frame(0, 8, 0)
	var p *pulse
	for _, candidate := range d.sticks[0] {
		if candidate.start > 2*ms && candidate.gap > 2*ms {
			p = candidate
			break
		}
	}
	if p == nil {
		t.Fatal("no LED with a start offset and a gap; change the seed")
	}
	start, gap := p.start, p.gap
	if got := p.at(start - ms); got != (blinkstick.RGB{}) {
		t.Errorf("before start: %v", got)
	}
	if got := p.at(start + 100*ms + gap/2); got != (blinkstick.RGB{}) {
		t.Errorf("during gap: %v", got)
	}
}

func TestDiscoTimingStaysInBounds(t *testing.T) {
	cfg := DiscoConfig{MinPeriod: 200 * ms, MaxPeriod: 2 * time.Second, MaxGap: time.Second, Colour: white, Rand: seeded(3)}
	d, _ := NewDisco(cfg)
	for at := time.Duration(0); at < 60*time.Second; at += 20 * ms {
		d.Frame(0, 8, at)
		for i, p := range d.sticks[0] {
			if p.period < cfg.MinPeriod || p.period > cfg.MaxPeriod || p.gap < 0 || p.gap > cfg.MaxGap {
				t.Fatalf("t=%v LED %d: period %v gap %v out of bounds", at, i, p.period, p.gap)
			}
			if at >= p.start+p.period+p.gap {
				t.Fatalf("t=%v LED %d: state not advanced", at, i)
			}
		}
	}
}

func TestDiscoLEDsAreIndependent(t *testing.T) {
	d, _ := NewDisco(DiscoConfig{MinPeriod: 200 * ms, MaxPeriod: 2 * time.Second, MaxGap: time.Second, Colour: white, Rand: seeded(4)})
	for at := time.Duration(0); at < 5*time.Second; at += 20 * ms {
		f := d.Frame(0, 8, at)
		if slices.ContainsFunc(f, func(c blinkstick.RGB) bool { return c != f[0] }) {
			return
		}
	}
	t.Error("every LED showed the same thing for 5 seconds")
}

func TestDiscoRepeatsWithTheSameSeed(t *testing.T) {
	colour := func(r *rand.Rand) blinkstick.RGB { return blinkstick.RGB{R: uint8(r.IntN(256))} }
	mk := func() *Disco {
		d, _ := NewDisco(DiscoConfig{MinPeriod: 200 * ms, MaxPeriod: 2 * time.Second, MaxGap: time.Second, Colour: colour, Rand: seeded(42)})
		return d
	}
	a, b := mk(), mk()
	for at := time.Duration(0); at < 5*time.Second; at += 20 * ms {
		if fa, fb := a.Frame(0, 8, at), b.Frame(0, 8, at); !slices.Equal(fa, fb) {
			t.Fatalf("t=%v: %v != %v", at, fa, fb)
		}
	}
}

func TestDiscoKeepsSticksSeparate(t *testing.T) {
	d, _ := NewDisco(DiscoConfig{MinPeriod: 200 * ms, MaxPeriod: 200 * ms, Colour: white, Rand: seeded(5)})
	d.Frame(0, 2, 0)
	d.Frame(1, 8, 0)
	if len(d.sticks[0]) != 2 || len(d.sticks[1]) != 8 {
		t.Errorf("stick states: %d and %d LEDs", len(d.sticks[0]), len(d.sticks[1]))
	}
}
