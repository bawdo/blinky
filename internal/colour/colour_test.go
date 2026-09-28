package colour

import (
	"errors"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/bawdo/go-blinkstick"

	"github.com/bawdo/blinky/internal/exitcode"
)

var (
	red  = blinkstick.RGB{R: 255}
	blue = blinkstick.RGB{B: 255}
)

func seeded(seed uint64) *rand.Rand { return rand.New(rand.NewPCG(seed, seed)) }

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want Spec
	}{
		{"ff0000", Spec{Kind: Fixed, RGB: red}},
		{"#0000ff", Spec{Kind: Fixed, RGB: blue}},
		{"f80", Spec{Kind: Fixed, RGB: blinkstick.RGB{R: 255, G: 136}}},
		{"255,136,0", Spec{Kind: Fixed, RGB: blinkstick.RGB{R: 255, G: 136}}},
		{"CornflowerBlue", Spec{Kind: Fixed, RGB: blinkstick.RGB{R: 100, G: 149, B: 237}}},
		{"off", Spec{Kind: Fixed}},
		{"OFF", Spec{Kind: Fixed}},
		{"random", Spec{Kind: Random}},
		{" Vivid ", Spec{Kind: Vivid}},
	}
	for _, tc := range cases {
		got, err := Parse(tc.in)
		if err != nil || got != tc.want {
			t.Errorf("Parse(%q) = %+v, %v; want %+v", tc.in, got, err, tc.want)
		}
	}
}

func TestParseRejectsBadColours(t *testing.T) {
	for _, in := range []string{"notacolour", "#12", "300,0,0", ""} {
		if _, err := Parse(in); !errors.Is(err, exitcode.ErrInvalidArgs) {
			t.Errorf("Parse(%q) error = %v, want ErrInvalidArgs", in, err)
		}
	}
}

func TestParseAllStopsAtFirstBadColour(t *testing.T) {
	if _, err := ParseAll([]string{"red", "nope"}); !errors.Is(err, exitcode.ErrInvalidArgs) {
		t.Errorf("error = %v, want ErrInvalidArgs", err)
	}
	specs, err := ParseAll([]string{"red", "blue"})
	if err != nil || len(specs) != 2 || specs[1].RGB != blue {
		t.Errorf("ParseAll = %+v, %v", specs, err)
	}
}

func TestIndex(t *testing.T) {
	cases := []struct{ i, n, leds, want int }{
		{0, 1, 8, 0}, {7, 1, 8, 0},
		{0, 2, 2, 0}, {1, 2, 2, 1},
		{3, 2, 8, 0}, {4, 2, 8, 1},
		{1, 4, 8, 0}, {2, 4, 8, 1}, {7, 4, 8, 3},
		{0, 8, 2, 0}, {1, 8, 2, 4},
	}
	for _, tc := range cases {
		if got := Index(tc.i, tc.n, tc.leds); got != tc.want {
			t.Errorf("Index(%d, %d, %d) = %d, want %d", tc.i, tc.n, tc.leds, got, tc.want)
		}
	}
}

func TestFrameSplitsEvenly(t *testing.T) {
	specs := []Spec{{RGB: red}, {RGB: blue}}
	if got := Frame(specs, 2, seeded(1)); !slices.Equal(got, []blinkstick.RGB{red, blue}) {
		t.Errorf("Nano frame = %v", got)
	}
	want := []blinkstick.RGB{red, red, red, red, blue, blue, blue, blue}
	if got := Frame(specs, 8, seeded(1)); !slices.Equal(got, want) {
		t.Errorf("Square frame = %v", got)
	}
}

func TestFramePicksOncePerGroup(t *testing.T) {
	got := Frame([]Spec{{Kind: Random}, {Kind: Vivid}}, 8, seeded(3))
	for i := 1; i < 4; i++ {
		if got[i] != got[0] || got[i+4] != got[4] {
			t.Fatalf("groups not uniform: %v", got)
		}
	}
}

func TestPickIsRepeatableWithASeed(t *testing.T) {
	a, b := seeded(9), seeded(9)
	for range 20 {
		if x, y := (Spec{Kind: Random}).Pick(a), (Spec{Kind: Random}).Pick(b); x != y {
			t.Fatalf("random picks differ: %v %v", x, y)
		}
	}
}

func TestVividIsFullySaturated(t *testing.T) {
	r := seeded(5)
	for range 500 {
		c := Spec{Kind: Vivid}.Pick(r)
		hi, lo := max(c.R, c.G, c.B), min(c.R, c.G, c.B)
		if hi != 255 || lo != 0 {
			t.Fatalf("vivid colour %v is not saturated", c)
		}
	}
	if vivid(0) != red || vivid(4*255) != blue {
		t.Errorf("vivid wheel: vivid(0)=%v vivid(1020)=%v", vivid(0), vivid(4*255))
	}
}

func TestSuggestions(t *testing.T) {
	s := Suggestions()
	if len(s) != 151 {
		t.Errorf("got %d suggestions, want 3 keywords + 148 names", len(s))
	}
	if s[0].Value != "off" || s[1].Value != "random" || s[2].Value != "vivid" {
		t.Errorf("keywords first, got %v", s[:3])
	}
	if !slices.Contains(s, Suggestion{Value: "cornflowerblue", Description: "#6495ed"}) {
		t.Error("cornflowerblue missing or without its hex")
	}
}
