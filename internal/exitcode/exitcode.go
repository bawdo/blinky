// Package exitcode owns the sentinel errors that map to documented CLI
// exit codes. Wrap a sentinel with fmt.Errorf("%w: %s", ErrXxx, detail)
// and cmd/root.go's classifyError will translate it to the right exit
// code on the way out.
//
// The set of codes is part of the CLI's public contract - keep it stable
// and document additions in README.md.
package exitcode

import "errors"

// Sentinel errors. Add new ones here AND extend the switch in From below
// AND document the new code in README.md.
var (
	ErrInvalidArgs  = errors.New("invalid arguments")
	ErrPrerequisite = errors.New("prerequisite missing")
	ErrNotFound     = errors.New("not found")
	ErrBusy         = errors.New("busy")
	ErrPartial      = errors.New("partial failure")
)

// From maps an error to its documented exit code:
//
//	0 - nil
//	1 - generic failure
//	2 - ErrInvalidArgs
//	3 - ErrPrerequisite
//	4 - ErrNotFound
//	5 - ErrBusy
//	6 - ErrPartial
//
// ErrPartial is checked first: a partial failure wraps the per-stick errors,
// which may match other sentinels, and must still exit 6.
func From(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, ErrPartial):
		return 6
	case errors.Is(err, ErrInvalidArgs):
		return 2
	case errors.Is(err, ErrPrerequisite):
		return 3
	case errors.Is(err, ErrNotFound):
		return 4
	case errors.Is(err, ErrBusy):
		return 5
	default:
		return 1
	}
}
