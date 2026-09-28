package cmd

import (
	"slices"
	"strings"
	"testing"

	"github.com/bawdo/go-blinkstick"
)

func TestColourCommandSetsAndReads(t *testing.T) {
	ctl := twoSticks()
	if res := run(t, opts(ctl), "colour", "-a", "red", "blue"); res.code != 0 {
		t.Fatalf("set: exit %d err %q", res.code, res.err)
	}
	if got := ctl.Stick("BS072777-3.0").Current(); got[1] != (blinkstick.RGB{B: 255}) {
		t.Errorf("Nano %v", got)
	}
	res := run(t, opts(ctl), "colour", "-d", "desk")
	if res.out != "desk: #ff0000 #0000ff\n" {
		t.Errorf("read %q", res.out)
	}
}

func TestColorIsAHiddenCopy(t *testing.T) {
	ctl := twoSticks()
	if res := run(t, opts(ctl), "color", "-d", "desk", "lime"); res.code != 0 {
		t.Fatalf("exit %d err %q", res.code, res.err)
	}
	if help := run(t, opts(ctl), "--help").out; strings.Contains(help, "\n  color ") || !strings.Contains(help, "\n  colour ") {
		t.Errorf("root help:\n%s", help)
	}
}

func TestColourCommandRejectsBadInput(t *testing.T) {
	cases := [][]string{
		{"colour", "-a", "notacolour"},
		{"colour", "-a", "--led", "-2", "red"},
		{"colour", "-a", "--json", "red"},
		{"colour", "-a", "--brightness", "101", "red"},
	}
	for _, args := range cases {
		if res := run(t, opts(twoSticks()), args...); res.code != 2 {
			t.Errorf("%v: exit %d, want 2", args, res.code)
		}
	}
}

func TestColourCompletion(t *testing.T) {
	if got := complete(t, opts(twoSticks()), "colour", "corn"); !slices.Equal(got, []string{"cornflowerblue", "cornsilk"}) {
		t.Errorf("got %v", got)
	}
	got := complete(t, opts(twoSticks()), "colour", "red", "o")
	if !slices.Contains(got, "off") || !slices.Contains(got, "orange") {
		t.Errorf("got %v", got)
	}
}

func TestOffCommand(t *testing.T) {
	ctl := twoSticks()
	run(t, opts(ctl), "colour", "-a", "red")
	if res := run(t, opts(ctl), "off", "-a"); res.code != 0 {
		t.Fatalf("exit %d err %q", res.code, res.err)
	}
	if got := ctl.Stick("BS073788-3.1").Current(); got[0] != (blinkstick.RGB{}) {
		t.Errorf("Square %v", got)
	}
}
