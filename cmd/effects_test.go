package cmd

import (
	"slices"
	"strings"
	"testing"

	"github.com/bawdo/go-blinkstick"
)

func TestBlinkCommand(t *testing.T) {
	ctl := twoSticks()
	res := run(t, opts(ctl), "blink", "-d", "desk", "red", "--period", "500ms", "--repeats", "2")
	if res.code != 0 {
		t.Fatalf("exit %d err %q", res.code, res.err)
	}
	frames := ctl.Stick("BS072777-3.0").Frames()
	if len(frames) != 51 { // two 500ms cycles every 20ms, then off
		t.Fatalf("got %d frames", len(frames))
	}
	if frames[12][0] != (blinkstick.RGB{R: 255}) || frames[13][0] != (blinkstick.RGB{}) || frames[50][0] != (blinkstick.RGB{}) {
		t.Errorf("frames 12, 13, 50: %v %v %v", frames[12], frames[13], frames[50])
	}
}

func TestBlinkOnOffAndSecondColourCommand(t *testing.T) {
	ctl := twoSticks()
	res := run(t, opts(ctl), "blink", "-d", "desk", "red", "--on-time", "100ms", "--off-time", "200ms",
		"--second-color", "yellow", "--repeats", "1")
	if res.code != 0 {
		t.Fatalf("exit %d err %q", res.code, res.err)
	}
	frames := ctl.Stick("BS072777-3.0").Frames()
	if len(frames) != 31 {
		t.Fatalf("got %d frames", len(frames))
	}
	if frames[15][0] != (blinkstick.RGB{R: 255, G: 255}) {
		t.Errorf("frame 15: %v", frames[15])
	}
}

func TestBlinkRejectsBadTimes(t *testing.T) {
	for _, args := range [][]string{
		{"blink", "-d", "desk", "red", "--on-time", "0s"},
		{"blink", "-d", "desk", "red", "--off-time", "-1s"},
		{"blink", "-d", "desk", "red", "--second-colour", "nope"},
	} {
		if res := run(t, opts(twoSticks()), args...); res.code != 2 {
			t.Errorf("%v: exit %d, want 2", args, res.code)
		}
	}
}

func TestColorSpellingIsHidden(t *testing.T) {
	help := run(t, opts(twoSticks()), "blink", "--help").out
	if !strings.Contains(help, "--second-colour") || strings.Contains(help, "second-color ") {
		t.Errorf("help:\n%s", help)
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

func TestMorphFromColourCommand(t *testing.T) {
	ctl := twoSticks()
	res := run(t, opts(ctl), "morph", "-d", "desk", "red", "--from-color", "lime", "--fade", "500ms")
	if res.code != 0 {
		t.Fatalf("exit %d err %q", res.code, res.err)
	}
	s := ctl.Stick("BS072777-3.0")
	if f := s.Frames(); len(f) == 0 || f[0][0] != (blinkstick.RGB{G: 255}) {
		t.Errorf("frames %v", f)
	}
	if !slices.Equal(s.Calls(), []string{"morph #ff0000 500ms"}) {
		t.Errorf("calls %v", s.Calls())
	}
}

func TestMorphLoopCommand(t *testing.T) {
	ctl := twoSticks()
	res := run(t, opts(ctl), "morph", "-d", "desk", "red", "--loop", "--fade", "100ms", "--repeats", "1")
	if res.code != 0 {
		t.Fatalf("exit %d err %q", res.code, res.err)
	}
	frames := ctl.Stick("BS072777-3.0").Frames()
	if len(frames) != 11 || frames[10][0] != (blinkstick.RGB{}) {
		t.Errorf("got %d frames, last %v", len(frames), frames[len(frames)-1])
	}
}

func TestMorphRejectsRepeatsWithoutLoop(t *testing.T) {
	if res := run(t, opts(twoSticks()), "morph", "-d", "desk", "red", "--repeats", "2"); res.code != 2 {
		t.Errorf("exit %d, want 2", res.code)
	}
}

func TestMorphHelpHidesColorSpelling(t *testing.T) {
	help := run(t, opts(twoSticks()), "morph", "--help").out
	for _, want := range []string{"--from-colour", "--fade", "--loop", "--repeats", "--duration"} {
		if !strings.Contains(help, want) {
			t.Errorf("help lacks %s:\n%s", want, help)
		}
	}
	if strings.Contains(help, "from-color ") {
		t.Errorf("help shows the color spelling:\n%s", help)
	}
}
