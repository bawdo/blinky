package cmd

import (
	"slices"
	"strings"
	"testing"
)

func TestPoliceCommand(t *testing.T) {
	ctl := twoSticks()
	if res := run(t, opts(ctl), "police", "-a", "--duration", "100ms", "--alternate"); res.code != 0 {
		t.Fatalf("exit %d err %q", res.code, res.err)
	}
	if n := len(ctl.Stick("BS072777-3.0").Frames()); n != 6 {
		t.Errorf("got %d frames", n)
	}
}

func TestDiscoCommand(t *testing.T) {
	ctl := twoSticks()
	if res := run(t, opts(ctl), "disco", "-a", "red", "gold", "--duration", "1s", "--seed", "3"); res.code != 0 {
		t.Fatalf("exit %d err %q", res.code, res.err)
	}
	if n := len(ctl.Stick("BS073788-3.1").Frames()); n != 51 { // 50 steps plus off
		t.Errorf("got %d frames", n)
	}
}

func TestModesRejectBadInput(t *testing.T) {
	cases := [][]string{
		{"police", "-a", "red"},
		{"disco", "-a", "notacolour"},
		{"disco", "-a", "--min-period", "0s"},
	}
	for _, args := range cases {
		if res := run(t, opts(twoSticks()), args...); res.code != 2 {
			t.Errorf("%v: exit %d, want 2", args, res.code)
		}
	}
}

func TestPoliceCompletesOnlyItsOwnFlags(t *testing.T) {
	got := complete(t, opts(twoSticks()), "police", "--")
	for _, want := range []string{"--alternate", "--period", "--duration", "--device"} {
		if !slices.Contains(got, want) {
			t.Errorf("missing %s in %v", want, got)
		}
	}
	if slices.Contains(got, "--min-period") {
		t.Errorf("offered a disco flag: %v", got)
	}
}

func TestDiscoCompletesColoursAndHidesSeed(t *testing.T) {
	if got := complete(t, opts(twoSticks()), "disco", "hotp"); !slices.Equal(got, []string{"hotpink"}) {
		t.Errorf("got %v", got)
	}
	if help := run(t, opts(twoSticks()), "disco", "--help").out; strings.Contains(help, "--seed") {
		t.Error("--seed shows in help")
	}
}
