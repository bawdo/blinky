// Package colour parses colour arguments and spreads colours over a
// stick's LEDs.
//
// Given N colours, a stick with L LEDs shows colour floor(i*N/L) on LED i,
// so two colours light the halves of a Square and one LED each of a Nano.
package colour

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/bawdo/go-blinkstick"

	"github.com/bawdo/blinky/internal/exitcode"
)

// Kind says how a Spec becomes a colour.
type Kind int

const (
	Fixed  Kind = iota // the RGB as given
	Random             // any colour
	Vivid              // a bright, fully saturated colour
)

// Keywords accepted on top of everything blinkstick.ParseRGB takes.
const (
	KeywordOff    = "off"
	KeywordRandom = "random"
	KeywordVivid  = "vivid"
)

// Spec is one parsed colour argument.
type Spec struct {
	Kind Kind
	RGB  blinkstick.RGB // used when Kind is Fixed
}

// Parse reads a colour argument: hex with or without "#", "r,g,b", a CSS
// colour name, or one of the keywords off, random and vivid.
func Parse(s string) (Spec, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case KeywordOff:
		return Spec{Kind: Fixed}, nil
	case KeywordRandom:
		return Spec{Kind: Random}, nil
	case KeywordVivid:
		return Spec{Kind: Vivid}, nil
	}
	c, err := blinkstick.ParseRGB(s)
	if err != nil {
		return Spec{}, fmt.Errorf("%w: colour %q: use hex (ff8800), r,g,b (255,136,0), a CSS name, off, random or vivid",
			exitcode.ErrInvalidArgs, s)
	}
	return Spec{Kind: Fixed, RGB: c}, nil
}

// ParseAll parses every argument, stopping at the first bad one.
func ParseAll(args []string) ([]Spec, error) {
	specs := make([]Spec, len(args))
	for i, a := range args {
		s, err := Parse(a)
		if err != nil {
			return nil, err
		}
		specs[i] = s
	}
	return specs, nil
}

// Pick returns the colour s stands for, drawing from r for Random and
// Vivid so a seeded r repeats exactly.
func (s Spec) Pick(r *rand.Rand) blinkstick.RGB {
	switch s.Kind {
	case Random:
		n := r.Uint32()
		return blinkstick.RGB{
			R: uint8(n >> 16), //nolint:gosec // G115: taking the low byte is the point
			G: uint8(n >> 8),  //nolint:gosec // G115: taking the low byte is the point
			B: uint8(n),       //nolint:gosec // G115: taking the low byte is the point
		}
	case Vivid:
		return vivid(r.IntN(6 * 255))
	}
	return s.RGB
}

// vivid returns the fully saturated colour at step h of the colour wheel,
// h in [0, 1530). It is the wheel blinkstick.RandomVivid uses, repeated
// here because RandomVivid draws from the global generator and cannot be
// seeded.
func vivid(h int) blinkstick.RGB {
	x := uint8(h % 255) //nolint:gosec // G115: h % 255 is 0 to 254
	switch h / 255 {
	case 0:
		return blinkstick.RGB{R: 255, G: x}
	case 1:
		return blinkstick.RGB{R: 255 - x, G: 255}
	case 2:
		return blinkstick.RGB{G: 255, B: x}
	case 3:
		return blinkstick.RGB{G: 255 - x, B: 255}
	case 4:
		return blinkstick.RGB{R: x, B: 255}
	default:
		return blinkstick.RGB{R: 255, B: 255 - x}
	}
}

// Index returns which of n colours LED i shows on a stick with leds LEDs.
func Index(i, n, leds int) int {
	return i * n / leds
}

// Frame picks each spec once, so random and vivid give one colour per LED
// group, and spreads the picks over leds LEDs.
func Frame(specs []Spec, leds int, r *rand.Rand) []blinkstick.RGB {
	picked := make([]blinkstick.RGB, len(specs))
	for i, s := range specs {
		picked[i] = s.Pick(r)
	}
	out := make([]blinkstick.RGB, leds)
	for i := range out {
		out[i] = picked[Index(i, len(specs), leds)]
	}
	return out
}

// Suggestion is one shell completion for a colour argument.
type Suggestion struct {
	Value, Description string
}

// Suggestions returns the keywords, then every CSS colour name with its hex.
func Suggestions() []Suggestion {
	out := []Suggestion{
		{KeywordOff, "LEDs off"},
		{KeywordRandom, "any colour, picked at random"},
		{KeywordVivid, "a bright colour, picked at random"},
	}
	for _, name := range blinkstick.ColourNames() {
		c, _ := blinkstick.ParseRGB(name)
		out = append(out, Suggestion{name, c.Hex()})
	}
	return out
}
