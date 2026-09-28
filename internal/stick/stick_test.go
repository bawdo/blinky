package stick

import (
	"testing"

	"github.com/bawdo/go-blinkstick"
)

func TestFromNamed(t *testing.T) {
	cases := []struct {
		name string
		in   blinkstick.NamedInfo
		want Info
	}{
		{"named nano",
			blinkstick.NamedInfo{Info: blinkstick.Info{Serial: "BS072777-3.0", Version: "3.0",
				Manufacturer: "Agile Innovative Ltd", Product: "BlinkStick", Model: blinkstick.Nano}, Name: "desk"},
			Info{Serial: "BS072777-3.0", Firmware: "3.0", Manufacturer: "Agile Innovative Ltd",
				Product: "BlinkStick", Model: "Nano", LEDs: 2, Name: "desk", Status: StatusOK}},
		{"busy square",
			blinkstick.NamedInfo{Info: blinkstick.Info{Serial: "BS073788-3.1", Model: blinkstick.Square}, Busy: true},
			Info{Serial: "BS073788-3.1", Model: "Square", LEDs: 8, Status: StatusBusy}},
		{"unknown model",
			blinkstick.NamedInfo{Info: blinkstick.Info{Serial: "BS000002-2.0", Model: blinkstick.Model{Name: "unknown"}}},
			Info{Serial: "BS000002-2.0", Model: "unknown", Status: StatusUnsupported}},
	}
	for _, tc := range cases {
		if got := fromNamed(tc.in); got != tc.want {
			t.Errorf("%s: got %+v, want %+v", tc.name, got, tc.want)
		}
	}
}

func TestIDPrefersName(t *testing.T) {
	if id := (Info{Serial: "BS1", Name: "desk"}).ID(); id != "desk" {
		t.Errorf("ID = %q", id)
	}
	if id := (Info{Serial: "BS1"}).ID(); id != "BS1" {
		t.Errorf("ID = %q", id)
	}
}

func TestSortBySerial(t *testing.T) {
	infos := []Info{{Serial: "BS3"}, {Serial: "BS1"}, {Serial: "BS2"}}
	SortBySerial(infos)
	if infos[0].Serial != "BS1" || infos[2].Serial != "BS3" {
		t.Errorf("got %v", infos)
	}
}
