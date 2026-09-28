package target

import (
	"strings"
	"testing"

	"github.com/bawdo/blinky/internal/exitcode"
	"github.com/bawdo/blinky/internal/stick"
	"github.com/bawdo/blinky/internal/stick/sticktest"
)

var (
	desk   = sticktest.Nano("BS072777-3.0", "desk")
	square = sticktest.Square("BS073788-3.1", "")
	busy   = sticktest.Busy(sticktest.Square("BS000001-3.0", "shelf"))
	odd    = sticktest.Unsupported("BS000002-2.0")
)

func serials(infos []stick.Info) string {
	var s []string
	for _, i := range infos {
		s = append(s, i.Serial)
	}
	return strings.Join(s, ",")
}

func TestResolve(t *testing.T) {
	cases := []struct {
		name   string
		sticks []stick.Info
		req    Request
		scope  Scope
		want   string // serials, comma separated
	}{
		{"only stick", []stick.Info{desk}, Request{}, Group, "BS072777-3.0"},
		{"only drivable stick", []stick.Info{desk, odd}, Request{}, Group, "BS072777-3.0"},
		{"all skips unsupported", []stick.Info{square, odd, desk}, Request{All: true}, Group, "BS072777-3.0,BS073788-3.1"},
		{"all includes busy", []stick.Info{desk, busy}, Request{All: true}, Group, "BS000001-3.0,BS072777-3.0"},
		{"by serial", []stick.Info{desk, square}, Request{Devices: []string{"BS073788-3.1"}}, Group, "BS073788-3.1"},
		{"by name", []stick.Info{desk, square}, Request{Devices: []string{"desk"}}, One, "BS072777-3.0"},
		{"sorted and deduplicated", []stick.Info{desk, square},
			Request{Devices: []string{"BS073788-3.1", "desk", "BS072777-3.0"}}, Group, "BS072777-3.0,BS073788-3.1"},
	}
	for _, tc := range cases {
		got, err := Resolve(tc.sticks, tc.req, tc.scope)
		if err != nil || serials(got) != tc.want {
			t.Errorf("%s: got %q, %v; want %q", tc.name, serials(got), err, tc.want)
		}
	}
}

func TestResolveSerialBeatsName(t *testing.T) {
	sneaky := sticktest.Nano("BS000009-3.0", "BS073788-3.1")
	got, err := Resolve([]stick.Info{sneaky, square}, Request{Devices: []string{"BS073788-3.1"}}, Group)
	if err != nil || serials(got) != "BS073788-3.1" {
		t.Errorf("got %q, %v", serials(got), err)
	}
}

func TestResolveErrors(t *testing.T) {
	twins := []stick.Info{sticktest.Nano("BS1", "desk"), sticktest.Nano("BS2", "desk")}
	cases := []struct {
		name     string
		sticks   []stick.Info
		req      Request
		scope    Scope
		code     int
		contains string
	}{
		{"none attached", nil, Request{}, Group, 4, "no BlinkSticks attached"},
		{"none attached with all", []stick.Info{odd}, Request{All: true}, Group, 4, "no BlinkSticks attached"},
		{"must choose", []stick.Info{desk, square}, Request{}, Group, 2, "desk, BS073788-3.1"},
		{"must choose one", []stick.Info{desk, square}, Request{}, One, 2, "choose one with --device:"},
		{"all and device", []stick.Info{desk}, Request{All: true, Devices: []string{"desk"}}, Group, 2, "not both"},
		{"one scope with all", []stick.Info{desk}, Request{All: true}, One, 2, "one stick at a time"},
		{"one scope with two", []stick.Info{desk, square}, Request{Devices: []string{"desk", "BS073788-3.1"}}, One, 2, "one stick at a time"},
		{"unknown", []stick.Info{desk}, Request{Devices: []string{"shelf"}}, Group, 4, "run blinky list"},
		{"unknown with busy", []stick.Info{desk, busy}, Request{Devices: []string{"shelf"}}, Group, 4,
			"shelf not found, 1 stick busy: BS000001-3.0, try its serial"},
		{"duplicate name", twins, Request{Devices: []string{"desk"}}, Group, 4, "name of 2 sticks (BS1, BS2)"},
		{"unsupported", []stick.Info{odd}, Request{Devices: []string{"BS000002-2.0"}}, Group, 2, "cannot drive"},
	}
	for _, tc := range cases {
		_, err := Resolve(tc.sticks, tc.req, tc.scope)
		if code := exitcode.From(err); code != tc.code || err == nil || !strings.Contains(err.Error(), tc.contains) {
			t.Errorf("%s: error %v (exit %d), want exit %d containing %q", tc.name, err, code, tc.code, tc.contains)
		}
	}
}

func TestResolveErrorsSanitiseSerials(t *testing.T) {
	dirty := sticktest.Busy(sticktest.Nano("BS\x1b[31m", "shelf"))
	_, err := Resolve([]stick.Info{desk, dirty}, Request{Devices: []string{"unknown"}}, Group)
	if err == nil || strings.Contains(err.Error(), "\x1b") {
		t.Errorf("busy list not sanitised: %v", err)
	}

	twinsDirty := []stick.Info{sticktest.Nano("BS\x1b[31m", "desk"), sticktest.Nano("BS2", "desk")}
	_, err = Resolve(twinsDirty, Request{Devices: []string{"desk"}}, Group)
	if err == nil || strings.Contains(err.Error(), "\x1b") {
		t.Errorf("duplicate-name list not sanitised: %v", err)
	}

	unsupportedDirty := sticktest.Unsupported("BS\x1b[31m")
	_, err = Resolve([]stick.Info{unsupportedDirty}, Request{Devices: []string{"BS\x1b[31m"}}, Group)
	if err == nil || strings.Contains(err.Error(), "\x1b") {
		t.Errorf("drivable error not sanitised: %v", err)
	}
}

func TestHelp(t *testing.T) {
	if got := Help(Group, Group); got != "Targets: one or more sticks" {
		t.Errorf("got %q", got)
	}
	if got := Help(Group, One); got != "Targets: reads one or more sticks, writes one stick" {
		t.Errorf("got %q", got)
	}
}
