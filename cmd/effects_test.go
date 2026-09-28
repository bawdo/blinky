package cmd

import (
	"slices"
	"testing"
)

func TestBlinkCommand(t *testing.T) {
	ctl := twoSticks()
	res := run(t, opts(ctl), "blink", "-d", "desk", "red", "--period", "500ms", "--repeats", "2")
	if res.code != 0 {
		t.Fatalf("exit %d err %q", res.code, res.err)
	}
	want := []string{"blink #ff0000 500ms 1", "blink #ff0000 500ms 1"}
	if got := ctl.Stick("BS072777-3.0").Calls(); !slices.Equal(got, want) {
		t.Errorf("calls %v", got)
	}
}

func TestPulseAndMorphDefaults(t *testing.T) {
	ctl := twoSticks()
	run(t, opts(ctl), "pulse", "-d", "desk", "lime")
	run(t, opts(ctl), "morph", "-d", "desk", "blue")
	want := []string{"pulse #00ff00 2s 1", "pulse #00ff00 2s 1", "pulse #00ff00 2s 1", "morph #0000ff 1s"}
	if got := ctl.Stick("BS072777-3.0").Calls(); !slices.Equal(got, want) {
		t.Errorf("calls %v", got)
	}
}

func TestEffectCommandsNeedOneColour(t *testing.T) {
	for _, args := range [][]string{{"blink", "-a"}, {"pulse", "-a", "red", "blue"}, {"morph", "-a"}} {
		if res := run(t, opts(twoSticks()), args...); res.code != 2 {
			t.Errorf("%v: exit %d, want 2", args, res.code)
		}
	}
}

func TestDurationCompletion(t *testing.T) {
	got := complete(t, opts(twoSticks()), "blink", "--period", "")
	if !slices.Equal(got, []string{"500ms", "1s", "5s", "30s"}) {
		t.Errorf("got %v", got)
	}
	if got := complete(t, opts(twoSticks()), "blink", "red", ""); len(got) != 0 {
		t.Errorf("offered a second colour: %v", got)
	}
}
