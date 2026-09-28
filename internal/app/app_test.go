package app

import (
	"bytes"
	"os"
	"testing"

	"github.com/bawdo/blinky/internal/effect"
	"github.com/bawdo/blinky/internal/stick"
)

func TestNewUsesStdout(t *testing.T) {
	a := New()
	if a.Out() != os.Stdout {
		t.Errorf("New().Out(): want os.Stdout, got %v", a.Out())
	}
}

func TestNewWithOptionsHonoursExplicitOut(t *testing.T) {
	var buf bytes.Buffer
	a := NewWithOptions(Options{Out: &buf})
	if a.Out() != &buf {
		t.Errorf("Out not honoured: got %p, want %p", a.Out(), &buf)
	}
}

func TestNewWithOptionsZeroValueFallsBackToStdout(t *testing.T) {
	a := NewWithOptions(Options{})
	if a.Out() != os.Stdout {
		t.Errorf("zero-valued Out should fall back to stdout, got %v", a.Out())
	}
}

func TestNewWithOptionsDefaults(t *testing.T) {
	a := NewWithOptions(Options{})
	if a.Err() != os.Stderr {
		t.Errorf("Err: want os.Stderr, got %v", a.Err())
	}
	if _, ok := a.ctl.(stick.Hardware); !ok {
		t.Errorf("Controller: want stick.Hardware, got %T", a.ctl)
	}
	if _, ok := a.clock.(effect.RealClock); !ok {
		t.Errorf("Clock: want effect.RealClock, got %T", a.clock)
	}
}
