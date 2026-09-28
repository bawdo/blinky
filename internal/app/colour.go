package app

import (
	"strings"

	"github.com/bawdo/go-blinkstick"

	"github.com/bawdo/blinky/internal/colour"
	"github.com/bawdo/blinky/internal/exitcode"
	"github.com/bawdo/blinky/internal/render"
	"github.com/bawdo/blinky/internal/settings"
	"github.com/bawdo/blinky/internal/stick"
	"github.com/bawdo/blinky/internal/target"
)

type coloursJSON struct {
	ID      string   `json:"id"`
	Serial  string   `json:"serial"`
	Colours []string `json:"colours"`
}

// ReadColour prints each stick's LEDs, or only LED led when led >= 0.
// set is applied so an inverted stick reads back what was written.
func (a *App) ReadColour(req target.Request, set settings.Settings, led int, asJSON bool) error {
	infos, _, err := a.resolve(req, target.Group)
	if err != nil {
		return err
	}
	if err := checkLED(infos, led); err != nil {
		return err
	}
	read := make([][]blinkstick.RGB, len(infos))
	runErr := a.each(infos, &set, func(o opened) error {
		leds, err := readLEDs(o.st, led)
		if err != nil {
			return libError(err)
		}
		read[o.pos] = leds
		return nil
	})
	rows := make([]coloursJSON, 0, len(infos))
	for i, leds := range read {
		if leds == nil {
			continue
		}
		hexes := make([]string, len(leds))
		for k, c := range leds {
			hexes[k] = c.Hex()
		}
		rows = append(rows, coloursJSON{ID: infos[i].ID(), Serial: infos[i].Serial, Colours: hexes})
	}
	if err := printRows(a, asJSON, rows, func(r coloursJSON) (string, string) {
		return r.ID, strings.Join(r.Colours, " ")
	}); err != nil {
		return err
	}
	return runErr
}

func readLEDs(st stick.Stick, led int) ([]blinkstick.RGB, error) {
	if led < 0 {
		return st.Frame()
	}
	c, err := st.LED(led)
	if err != nil {
		return nil, err
	}
	return []blinkstick.RGB{c}, nil
}

// SetColour spreads specs over every chosen stick's LEDs, or sets only LED
// led when led >= 0, which takes exactly one colour.
func (a *App) SetColour(req target.Request, set settings.Settings, led int, specs []colour.Spec) error {
	switch {
	case len(specs) == 0:
		return exitcode.Invalid("give at least one colour")
	case led >= 0 && len(specs) != 1:
		return exitcode.Invalid("--led takes exactly one colour")
	}
	infos, _, err := a.resolve(req, target.Group)
	if err != nil {
		return err
	}
	if err := checkLED(infos, led); err != nil {
		return err
	}
	// Pick every colour up front: *rand.Rand is not safe across the
	// goroutines each starts.
	r := newRand(0)
	frames := make([][]blinkstick.RGB, len(infos))
	for i, info := range infos {
		if led >= 0 {
			frames[i] = []blinkstick.RGB{specs[0].Pick(r)}
		} else {
			frames[i] = colour.Frame(specs, info.LEDs, r)
		}
	}
	return a.each(infos, &set, func(o opened) error {
		if led >= 0 {
			return libError(o.st.SetLED(led, frames[o.pos][0]))
		}
		return libError(o.st.SetFrame(frames[o.pos]))
	})
}

// Off turns every LED off on the chosen sticks.
func (a *App) Off(req target.Request, set settings.Settings) error {
	infos, _, err := a.resolve(req, target.Group)
	if err != nil {
		return err
	}
	return a.each(infos, &set, func(o opened) error { return libError(o.st.Off()) })
}

// checkLED makes sure LED led exists on every stick. led < 0 means all.
func checkLED(infos []stick.Info, led int) error {
	if led < 0 {
		return nil
	}
	for _, i := range infos {
		if led >= i.LEDs {
			return exitcode.Invalid("LED %d does not exist on %s (a %s has LEDs 0 to %d)",
				led, render.Sanitise(i.ID()), i.Model, i.LEDs-1)
		}
	}
	return nil
}
