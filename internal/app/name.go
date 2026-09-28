package app

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/bawdo/blinky/internal/exitcode"
	"github.com/bawdo/blinky/internal/render"
	"github.com/bawdo/blinky/internal/stick"
	"github.com/bawdo/blinky/internal/target"
)

// maxBlock is the size of an info block, and so the longest name.
const maxBlock = 32

// serialShape matches BlinkStick serials such as BS012345-3.0. A name like
// that would be confused with a serial by --device.
var serialShape = regexp.MustCompile(`^BS\d+-\d+\.\d+$`)

type nameJSON struct {
	ID     string `json:"id"`
	Serial string `json:"serial"`
	Name   string `json:"name"`
}

// ReadName prints the name stored on each chosen stick.
func (a *App) ReadName(req target.Request, asJSON bool) error {
	infos, _, err := a.resolve(req, target.Group)
	if err != nil {
		return err
	}
	names := make([]*string, len(infos))
	runErr := a.each(infos, nil, func(o opened) error {
		n, err := o.st.Name()
		if err != nil {
			return libError(err)
		}
		names[o.pos] = &n
		return nil
	})
	rows := make([]nameJSON, 0, len(infos))
	for i, n := range names {
		if n != nil {
			rows = append(rows, nameJSON{ID: infos[i].ID(), Serial: infos[i].Serial, Name: *n})
		}
	}
	if err := printRows(a, asJSON, rows, func(r nameJSON) (string, string) {
		if r.Name == "" {
			return r.ID, "-"
		}
		return r.ID, render.Sanitise(r.Name)
	}); err != nil {
		return err
	}
	return runErr
}

// SetName stores name on one stick. "" clears it.
func (a *App) SetName(req target.Request, name string) error {
	infos, all, err := a.resolve(req, target.One)
	if err != nil {
		return err
	}
	if name != "" {
		if err := validateName(name, infos[0], all); err != nil {
			return err
		}
	}
	return a.each(infos, nil, func(o opened) error { return libError(o.st.SetName(name)) })
}

// validateName checks name for self against every attached stick. Busy
// sticks' names cannot be read, so duplicates on them are not caught.
func validateName(name string, self stick.Info, all []stick.Info) error {
	switch {
	case !utf8.ValidString(name):
		return exitcode.Invalid("names must be UTF-8")
	case strings.ContainsFunc(name, unicode.IsControl):
		return exitcode.Invalid("name %s contains control characters", render.Sanitise(name))
	case len(name) > maxBlock:
		return exitcode.Invalid("name is %d bytes, a stick holds at most %d", len(name), maxBlock)
	case serialShape.MatchString(name):
		return exitcode.Invalid("%s looks like a serial, which would confuse --device", name)
	}
	for _, other := range all {
		if other.Serial != self.Serial && other.Name == name {
			return exitcode.Invalid("%s is already named %s, names must be unique", render.Sanitise(other.Serial), render.Sanitise(name))
		}
	}
	return nil
}
