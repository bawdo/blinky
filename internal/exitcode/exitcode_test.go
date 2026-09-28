package exitcode

import (
	"errors"
	"fmt"
	"testing"
)

func TestFromMapping(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, 0},
		{"generic", errors.New("boom"), 1},
		{"invalid args", fmt.Errorf("wrap: %w", ErrInvalidArgs), 2},
		{"prerequisite", fmt.Errorf("wrap: %w", ErrPrerequisite), 3},
		{"double-wrapped invalid args", fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", ErrInvalidArgs)), 2},
		{"not found", fmt.Errorf("wrap: %w", ErrNotFound), 4},
		{"busy", fmt.Errorf("wrap: %w", ErrBusy), 5},
		{"partial", fmt.Errorf("wrap: %w", ErrPartial), 6},
		{"partial wins over what it wraps", errors.Join(ErrPartial, fmt.Errorf("desk: %w", ErrInvalidArgs)), 6},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := From(tc.err)
			if got != tc.want {
				t.Errorf("From(%v) = %d, want %d", tc.err, got, tc.want)
			}
		})
	}
}
