package settings

import (
	"errors"
	"testing"

	"github.com/bawdo/blinky/internal/exitcode"
)

func ptr[T any](v T) *T { return &v }

func TestResolveDefaults(t *testing.T) {
	got, err := Resolve()
	if err != nil || got != (Settings{Brightness: 100}) {
		t.Errorf("Resolve() = %+v, %v", got, err)
	}
}

func TestResolveLaterLayersWin(t *testing.T) {
	config := Layer{Brightness: ptr(30), Inverse: ptr(true)}
	flags := Layer{Brightness: ptr(80)}
	got, err := Resolve(config, flags)
	if err != nil || got != (Settings{Brightness: 80, Inverse: true}) {
		t.Errorf("Resolve(config, flags) = %+v, %v", got, err)
	}
}

func TestResolveRejectsBrightnessOutOfRange(t *testing.T) {
	for _, b := range []int{-1, 101} {
		if _, err := Resolve(Layer{Brightness: ptr(b)}); !errors.Is(err, exitcode.ErrInvalidArgs) {
			t.Errorf("brightness %d: error = %v, want ErrInvalidArgs", b, err)
		}
	}
}

func TestLimit(t *testing.T) {
	cases := map[int]uint8{0: 0, 1: 3, 40: 102, 50: 128, 100: 255}
	for pct, want := range cases {
		if got := (Settings{Brightness: pct}).Limit(); got != want {
			t.Errorf("Limit(%d%%) = %d, want %d", pct, got, want)
		}
	}
}

type recorder struct {
	limit   uint8
	inverse bool
}

func (r *recorder) SetBrightnessLimit(l uint8) { r.limit = l }
func (r *recorder) SetInverse(on bool)         { r.inverse = on }

func TestApply(t *testing.T) {
	var r recorder
	Settings{Brightness: 40, Inverse: true}.Apply(&r)
	if r.limit != 102 || !r.inverse {
		t.Errorf("applied %+v", r)
	}
}
