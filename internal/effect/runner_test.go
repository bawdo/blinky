package effect

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/bawdo/go-blinkstick"

	"github.com/bawdo/blinky/internal/effect/effecttest"
)

const ms = time.Millisecond

// clockEffect puts elapsed milliseconds in R and the sink index in G.
type clockEffect struct{}

func (clockEffect) Frame(sink, leds int, t time.Duration) []blinkstick.RGB {
	out := make([]blinkstick.RGB, leds)
	for i := range out {
		out[i] = blinkstick.RGB{R: uint8(t / ms), G: uint8(sink)}
	}
	return out
}

type recordSink struct {
	leds  int
	clock *effecttest.Clock
	fail  func(t time.Duration) error
	at    []time.Duration // successful writes
	tried []time.Duration // every attempt
	last  []blinkstick.RGB
}

func (s *recordSink) LEDs() int { return s.leds }

func (s *recordSink) SetFrame(leds []blinkstick.RGB) error {
	t := s.clock.Elapsed()
	s.tried = append(s.tried, t)
	if s.fail != nil {
		if err := s.fail(t); err != nil {
			return err
		}
	}
	s.at = append(s.at, t)
	s.last = leds
	return nil
}

func TestRunnerWritesEveryStepUntilDuration(t *testing.T) {
	c := effecttest.NewClock()
	a, b := &recordSink{leds: 2, clock: c}, &recordSink{leds: 8, clock: c}
	errs := Runner{Step: 20 * ms, Duration: 100 * ms, Clock: c}.Run(context.Background(), clockEffect{}, []Sink{a, b})
	if errs[0] != nil || errs[1] != nil {
		t.Fatalf("errs = %v", errs)
	}
	if want := []time.Duration{0, 20 * ms, 40 * ms, 60 * ms, 80 * ms}; !slices.Equal(a.at, want) {
		t.Errorf("writes at %v, want %v", a.at, want)
	}
	if a.last[0].R != 80 || len(b.last) != 8 || b.last[0].G != 1 {
		t.Errorf("last frames a=%v b=%v", a.last, b.last)
	}
}

func TestRunnerDefaultsToTwentyMillisecondSteps(t *testing.T) {
	c := effecttest.NewClock()
	a := &recordSink{leds: 2, clock: c}
	Runner{Duration: 50 * ms, Clock: c}.Run(context.Background(), clockEffect{}, []Sink{a})
	if want := []time.Duration{0, 20 * ms, 40 * ms}; !slices.Equal(a.at, want) {
		t.Errorf("writes at %v, want %v", a.at, want)
	}
}

func TestRunnerStopsWhenContextEnds(t *testing.T) {
	c := effecttest.NewClock()
	a := &recordSink{leds: 2, clock: c}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errs := Runner{Clock: c}.Run(ctx, clockEffect{}, []Sink{a})
	if errs[0] != nil || len(a.tried) != 0 {
		t.Errorf("errs=%v tried=%v", errs, a.tried)
	}
}

func TestRunnerRetriesDisconnectedSinks(t *testing.T) {
	c := effecttest.NewClock()
	a := &recordSink{leds: 2, clock: c, fail: func(t time.Duration) error {
		if t >= 20*ms && t < 60*ms {
			return blinkstick.ErrDisconnected
		}
		return nil
	}}
	var events []Event
	r := Runner{Step: 20 * ms, Retry: 50 * ms, Duration: 120 * ms, Clock: c, OnEvent: func(e Event) { events = append(events, e) }}
	errs := r.Run(context.Background(), clockEffect{}, []Sink{a})
	if errs[0] != nil {
		t.Fatalf("errs = %v", errs)
	}
	if want := []time.Duration{0, 20 * ms, 80 * ms, 100 * ms}; !slices.Equal(a.tried, want) {
		t.Errorf("tried at %v, want %v", a.tried, want)
	}
	if len(events) != 2 || events[0].Kind != Dropped || events[1].Kind != Returned ||
		!errors.Is(events[0].Err, blinkstick.ErrDisconnected) {
		t.Errorf("events = %+v", events)
	}
}

func TestRunnerDropsSinkOnOtherErrors(t *testing.T) {
	c := effecttest.NewClock()
	boom := errors.New("boom")
	a := &recordSink{leds: 2, clock: c, fail: func(time.Duration) error { return boom }}
	b := &recordSink{leds: 2, clock: c}
	errs := Runner{Duration: 60 * ms, Clock: c}.Run(context.Background(), clockEffect{}, []Sink{a, b})
	if !errors.Is(errs[0], boom) || errs[1] != nil {
		t.Errorf("errs = %v", errs)
	}
	if len(a.tried) != 1 || len(b.at) != 3 {
		t.Errorf("a tried %v, b wrote %v", a.tried, b.at)
	}
}

func TestRunnerReturnsEarlyWhenEverySinkFails(t *testing.T) {
	c := effecttest.NewClock()
	a := &recordSink{leds: 2, clock: c, fail: func(time.Duration) error { return errors.New("boom") }}
	errs := Runner{Clock: c}.Run(context.Background(), clockEffect{}, []Sink{a})
	if errs[0] == nil || c.Elapsed() != 0 {
		t.Errorf("errs=%v elapsed=%v", errs, c.Elapsed())
	}
}
