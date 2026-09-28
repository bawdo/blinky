package app

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/bawdo/blinky/internal/stick/sticktest"
	"github.com/bawdo/blinky/internal/target"
)

func mixed() *sticktest.Controller {
	return sticktest.New(
		sticktest.Nano("BS072777-3.0", "desk"),
		sticktest.Square("BS073788-3.1", ""),
		sticktest.Busy(sticktest.Square("BS000001-3.0", "")),
		sticktest.Unsupported("BS000002-2.0"),
	)
}

// line builds one table line from cells and column widths, the way
// render.Table pads.
func line(widths []int, cells ...string) string {
	var b strings.Builder
	for i, c := range cells[:len(cells)-1] {
		b.WriteString(c + strings.Repeat(" ", widths[i]-len(c)+2))
	}
	return b.String() + cells[len(cells)-1] + "\n"
}

func TestListTable(t *testing.T) {
	a, out, _ := newTestApp(mixed())
	if err := a.List(false); err != nil {
		t.Fatal(err)
	}
	w := []int{12, 4, 12, 7, 4}
	want := line(w, "ID", "NAME", "SERIAL", "MODEL", "LEDS", "STATUS") +
		line(w, "BS000001-3.0", "?", "BS000001-3.0", "Square", "8", "busy") +
		line(w, "BS000002-2.0", "-", "BS000002-2.0", "unknown", "-", "unsupported") +
		line(w, "desk", "desk", "BS072777-3.0", "Nano", "2", "ok") +
		line(w, "BS073788-3.1", "-", "BS073788-3.1", "Square", "8", "ok")
	if out.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", out.String(), want)
	}
}

func TestListWithNoSticks(t *testing.T) {
	a, out, errb := newTestApp(sticktest.New())
	if err := a.List(false); err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 || errb.String() != "No BlinkSticks attached.\n" {
		t.Errorf("out %q err %q", out.String(), errb.String())
	}
	if err := a.List(true); err != nil || out.String() != "[]\n" {
		t.Errorf("JSON: %q, %v", out.String(), err)
	}
}

func TestListJSON(t *testing.T) {
	a, out, _ := newTestApp(mixed())
	if err := a.List(true); err != nil {
		t.Fatal(err)
	}
	var got []map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	desk := got[2]
	if desk["id"] != "desk" || desk["serial"] != "BS072777-3.0" || desk["leds"] != 2.0 ||
		desk["firmware"] != "3.0" || desk["status"] != "ok" || desk["product"] != "BlinkStick Nano" {
		t.Errorf("desk = %v", desk)
	}
}

// field builds one info line: labels pad to "Manufacturer:" plus two.
func field(label, value string) string {
	return "  " + label + ":" + strings.Repeat(" ", 13-len(label)-1+2) + value + "\n"
}

func TestInfo(t *testing.T) {
	a, out, _ := newTestApp(mixed())
	if err := a.Info(target.Request{Devices: []string{"desk", "BS000001-3.0"}}, false); err != nil {
		t.Fatal(err)
	}
	want := "BS000001-3.0\n" +
		field("Serial", "BS000001-3.0") + field("Model", "Square (8 LEDs)") + field("Firmware", "3.1") +
		field("Manufacturer", "Agile Innovative Ltd") + field("Product", "BlinkStick Square") +
		field("Name", "?") + field("Status", "busy") +
		"\n" +
		"desk\n" +
		field("Serial", "BS072777-3.0") + field("Model", "Nano (2 LEDs)") + field("Firmware", "3.0") +
		field("Manufacturer", "Agile Innovative Ltd") + field("Product", "BlinkStick Nano") +
		field("Name", "desk") + field("Status", "ok")
	if out.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", out.String(), want)
	}
}
