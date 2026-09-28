package app

import (
	"strings"
	"testing"

	"github.com/bawdo/blinky/internal/exitcode"
	"github.com/bawdo/blinky/internal/stick/sticktest"
	"github.com/bawdo/blinky/internal/target"
)

func dev(id string) target.Request { return target.Request{Devices: []string{id}} }

func TestReadName(t *testing.T) {
	a, out, _ := newTestApp(deskAndSquare())
	if err := a.ReadName(all, false); err != nil {
		t.Fatal(err)
	}
	if want := "desk: desk\nBS073788-3.1: -\n"; out.String() != want {
		t.Errorf("got %q", out.String())
	}
}

func TestSetName(t *testing.T) {
	ctl := deskAndSquare()
	a, _, _ := newTestApp(ctl)
	if err := a.SetName(dev("BS073788-3.1"), "shelf"); err != nil {
		t.Fatal(err)
	}
	if got := ctl.Stick("BS073788-3.1").StoredName(); got != "shelf" {
		t.Errorf("stored %q", got)
	}
}

func TestSetNameKeepsTheSameNameOnTheSameStick(t *testing.T) {
	a, _, _ := newTestApp(deskAndSquare())
	if err := a.SetName(dev("desk"), "desk"); err != nil {
		t.Errorf("err %v", err)
	}
}

func TestClearName(t *testing.T) {
	ctl := deskAndSquare()
	a, _, _ := newTestApp(ctl)
	if err := a.SetName(dev("desk"), ""); err != nil {
		t.Fatal(err)
	}
	if got := ctl.Stick("BS072777-3.0").StoredName(); got != "" {
		t.Errorf("stored %q", got)
	}
}

func TestSetNameRefuses(t *testing.T) {
	cases := []struct {
		name, req, value, contains string
		all                        bool
	}{
		{"group", "", "x", "one stick at a time", true},
		{"duplicate", "BS073788-3.1", "desk", "already named", false},
		{"serial shaped", "BS073788-3.1", "BS012345-3.0", "looks like a serial", false},
		{"control characters", "BS073788-3.1", "a\tb", `a\tb contains control characters`, false},
		{"too long", "BS073788-3.1", strings.Repeat("x", 33), "33 bytes", false},
		{"not UTF-8", "BS073788-3.1", "\xff", "UTF-8", false},
	}
	for _, tc := range cases {
		a, _, _ := newTestApp(deskAndSquare())
		req := dev(tc.req)
		if tc.all {
			req = all
		}
		err := a.SetName(req, tc.value)
		if exitcode.From(err) != 2 || !strings.Contains(err.Error(), tc.contains) {
			t.Errorf("%s: err %v", tc.name, err)
		}
	}
}

func TestSetNameDuplicateErrorSanitisesOtherSerial(t *testing.T) {
	ctl := sticktest.New(sticktest.Nano("BS\x1b[31m", "desk"), sticktest.Nano("BS2", ""))
	a, _, _ := newTestApp(ctl)
	err := a.SetName(dev("BS2"), "desk")
	if exitcode.From(err) != 2 || strings.Contains(err.Error(), "\x1b") {
		t.Errorf("err %v", err)
	}
}

func TestSetNameOnABusyStick(t *testing.T) {
	a, _, _ := newTestApp(sticktest.New(sticktest.Busy(sticktest.Nano("BS1", ""))))
	if err := a.SetName(dev("BS1"), "desk"); exitcode.From(err) != 5 {
		t.Errorf("err %v", err)
	}
}
