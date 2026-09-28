// Package target turns --device and --all into the sticks a command acts
// on. It works on a stick list and never touches hardware.
package target

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bawdo/blinky/internal/exitcode"
	"github.com/bawdo/blinky/internal/render"
	"github.com/bawdo/blinky/internal/stick"
)

// Scope is how many sticks a command may act on.
type Scope int

const (
	One   Scope = iota // exactly one stick
	Group              // one or more sticks
)

// Noun describes s for help text.
func (s Scope) Noun() string {
	if s == One {
		return "one stick"
	}
	return "one or more sticks"
}

// Help is the "Targets:" line of a command's help.
func Help(read, write Scope) string {
	if read == write {
		return "Targets: " + read.Noun()
	}
	return "Targets: reads " + read.Noun() + ", writes " + write.Noun()
}

// Request is what the user asked for with --device and --all.
type Request struct {
	Devices []string // serials or names
	All     bool
}

// Resolve picks the sticks req names from sticks. With neither --device
// nor --all it picks the only stick blinky can drive. The result is sorted
// by serial.
func Resolve(sticks []stick.Info, req Request, scope Scope) ([]stick.Info, error) {
	if req.All && len(req.Devices) > 0 {
		return nil, fmt.Errorf("%w: use --device or --all, not both", exitcode.ErrInvalidArgs)
	}
	if scope == One && (req.All || len(req.Devices) > 1) {
		return nil, fmt.Errorf("%w: this works on one stick at a time, choose it with a single --device", exitcode.ErrInvalidArgs)
	}
	if len(req.Devices) == 0 {
		return resolveImplicit(sticks, req.All, scope)
	}
	var chosen []stick.Info
	for _, id := range req.Devices {
		s, err := find(sticks, id)
		if err != nil {
			return nil, err
		}
		if !slices.ContainsFunc(chosen, func(c stick.Info) bool { return c.Serial == s.Serial }) {
			chosen = append(chosen, s)
		}
	}
	stick.SortBySerial(chosen)
	return chosen, nil
}

func resolveImplicit(sticks []stick.Info, all bool, scope Scope) ([]stick.Info, error) {
	var usable []stick.Info
	for _, s := range sticks {
		if s.Status != stick.StatusUnsupported {
			usable = append(usable, s)
		}
	}
	stick.SortBySerial(usable)
	switch {
	case len(usable) == 0:
		return nil, fmt.Errorf("%w: no BlinkSticks attached", exitcode.ErrNotFound)
	case all, len(usable) == 1:
		return usable, nil
	}
	hint := "choose one with --device or use --all"
	if scope == One {
		hint = "choose one with --device"
	}
	ids := make([]string, len(usable))
	for i, s := range usable {
		ids[i] = render.Sanitise(s.ID())
	}
	return nil, fmt.Errorf("%w: %d sticks attached, %s: %s", exitcode.ErrInvalidArgs, len(usable), hint, strings.Join(ids, ", "))
}

// find matches id against serials first, then names.
func find(sticks []stick.Info, id string) (stick.Info, error) {
	for _, s := range sticks {
		if s.Serial == id {
			return drivable(s)
		}
	}
	var named []stick.Info
	for _, s := range sticks {
		if s.Name != "" && s.Name == id {
			named = append(named, s)
		}
	}
	shown := render.Sanitise(id)
	switch len(named) {
	case 1:
		return drivable(named[0])
	case 0:
		var busy []string
		for _, s := range sticks {
			if s.Status == stick.StatusBusy {
				busy = append(busy, render.Sanitise(s.Serial))
			}
		}
		if len(busy) > 0 {
			return stick.Info{}, fmt.Errorf("%w: %s not found, %s busy: %s, try its serial",
				exitcode.ErrNotFound, shown, count(len(busy)), strings.Join(busy, ", "))
		}
		return stick.Info{}, fmt.Errorf("%w: no stick has the serial or name %s, run blinky list to see them",
			exitcode.ErrNotFound, shown)
	}
	serials := make([]string, len(named))
	for i, s := range named {
		serials[i] = render.Sanitise(s.Serial)
	}
	return stick.Info{}, fmt.Errorf("%w: %s is the name of %d sticks (%s), use a serial",
		exitcode.ErrNotFound, shown, len(named), strings.Join(serials, ", "))
}

func drivable(s stick.Info) (stick.Info, error) {
	if s.Status == stick.StatusUnsupported {
		return stick.Info{}, fmt.Errorf("%w: %s is a model blinky cannot drive", exitcode.ErrInvalidArgs, render.Sanitise(s.Serial))
	}
	return s, nil
}

func count(n int) string {
	if n == 1 {
		return "1 stick"
	}
	return fmt.Sprintf("%d sticks", n)
}
