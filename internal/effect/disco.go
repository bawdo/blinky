package effect

import (
	"errors"
	"math/rand/v2"
	"time"

	"github.com/bawdo/go-blinkstick"
)

// DiscoConfig sets up a Disco.
type DiscoConfig struct {
	MinPeriod, MaxPeriod time.Duration                     // bounds of one pulse
	MaxGap               time.Duration                     // longest dark gap between pulses
	Colour               func(r *rand.Rand) blinkstick.RGB // picks each pulse's colour
	Rand                 *rand.Rand                        // every random choice comes from here
}

// Disco pulses every LED on every stick independently: a random colour,
// fading up and down over a random period, then dark for a random gap. It
// is not safe for concurrent use; Runner calls it from one goroutine.
type Disco struct {
	cfg    DiscoConfig
	sticks map[int][]*pulse
}

// pulse is one LED's current cycle: dark until start, lit for period, then
// dark for gap.
type pulse struct {
	colour             blinkstick.RGB
	start, period, gap time.Duration
}

// NewDisco checks cfg and returns a Disco.
func NewDisco(cfg DiscoConfig) (*Disco, error) {
	switch {
	case cfg.MinPeriod <= 0:
		return nil, errors.New("disco: min period must be more than 0")
	case cfg.MaxPeriod < cfg.MinPeriod:
		return nil, errors.New("disco: max period must not be less than min period")
	case cfg.MaxGap < 0:
		return nil, errors.New("disco: max gap must not be negative")
	case cfg.Colour == nil || cfg.Rand == nil:
		return nil, errors.New("disco: Colour and Rand are required")
	}
	return &Disco{cfg: cfg, sticks: map[int][]*pulse{}}, nil
}

// Frame implements Effect.
func (d *Disco) Frame(sink, leds int, t time.Duration) []blinkstick.RGB {
	pulses := d.pulses(sink, leds)
	out := make([]blinkstick.RGB, leds)
	for i, p := range pulses {
		for t >= p.start+p.period+p.gap {
			p.start += p.period + p.gap
			d.pick(p)
		}
		out[i] = p.at(t)
	}
	return out
}

// pulses returns sink's LED states, creating them on first use. Each LED
// starts after a random gap so they do not all begin together.
func (d *Disco) pulses(sink, leds int) []*pulse {
	if ps, ok := d.sticks[sink]; ok && len(ps) == leds {
		return ps
	}
	ps := make([]*pulse, leds)
	for i := range ps {
		ps[i] = &pulse{start: d.between(0, d.cfg.MaxGap)}
		d.pick(ps[i])
	}
	d.sticks[sink] = ps
	return ps
}

func (d *Disco) pick(p *pulse) {
	p.colour = d.cfg.Colour(d.cfg.Rand)
	p.period = d.between(d.cfg.MinPeriod, d.cfg.MaxPeriod)
	p.gap = d.between(0, d.cfg.MaxGap)
}

// between returns a random duration in [lo, hi].
func (d *Disco) between(lo, hi time.Duration) time.Duration {
	if hi <= lo {
		return lo
	}
	return lo + time.Duration(d.cfg.Rand.Int64N(int64(hi-lo)+1))
}

// at returns the LED's colour at t: dark outside the pulse, fading up to
// full at the middle and back down.
func (p *pulse) at(t time.Duration) blinkstick.RGB {
	x := t - p.start
	if x < 0 || x >= p.period {
		return blinkstick.RGB{}
	}
	half := p.period / 2
	if x < half {
		return dim(p.colour, ramp(x, half))
	}
	return dim(p.colour, ramp(p.period-x, p.period-half))
}
