package cmd

import (
	"slices"
	"strings"
	"testing"

	"github.com/bawdo/blinky/internal/stick/sticktest"
)

func twoSticks() *sticktest.Controller {
	return sticktest.New(sticktest.Nano("BS072777-3.0", "desk"), sticktest.Square("BS073788-3.1", ""),
		sticktest.Unsupported("BS000002-2.0"))
}

func TestListCommand(t *testing.T) {
	res := run(t, opts(twoSticks()), "list")
	if res.code != 0 || !strings.HasPrefix(res.out, "ID ") || !strings.Contains(res.out, "desk") {
		t.Errorf("exit %d out %q err %q", res.code, res.out, res.err)
	}
}

func TestListRejectsArguments(t *testing.T) {
	if res := run(t, opts(twoSticks()), "list", "extra"); res.code != 2 {
		t.Errorf("exit %d, want 2", res.code)
	}
}

func TestUnknownFlagExitsTwo(t *testing.T) {
	if res := run(t, opts(twoSticks()), "list", "--nope"); res.code != 2 {
		t.Errorf("exit %d, want 2", res.code)
	}
}

func TestInfoNeedsAChoiceWithTwoSticks(t *testing.T) {
	res := run(t, opts(twoSticks()), "info")
	if res.code != 2 || !strings.Contains(res.err, "desk, BS073788-3.1") {
		t.Errorf("exit %d err %q", res.code, res.err)
	}
}

func TestInfoByName(t *testing.T) {
	res := run(t, opts(twoSticks()), "info", "-d", "desk")
	if res.code != 0 || !strings.HasPrefix(res.out, "desk\n  Serial:") {
		t.Errorf("exit %d out %q", res.code, res.out)
	}
}

func TestInfoHelpShowsTargets(t *testing.T) {
	res := run(t, opts(twoSticks()), "info", "--help")
	if !strings.Contains(res.out, "Targets: one or more sticks") {
		t.Errorf("help: %s", res.out)
	}
}

func TestDeviceCompletion(t *testing.T) {
	if got := complete(t, opts(twoSticks()), "info", "-d", ""); !slices.Equal(got, []string{"desk", "BS073788-3.1"}) {
		t.Errorf("got %v", got)
	}
	if got := complete(t, opts(twoSticks()), "info", "-d", "de"); !slices.Equal(got, []string{"desk"}) {
		t.Errorf("got %v", got)
	}
}
