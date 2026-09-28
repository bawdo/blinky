package app

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/bawdo/blinky/internal/exitcode"
)

func TestReadInfoBlock(t *testing.T) {
	ctl := deskAndSquare()
	st, _ := ctl.Open("BS072777-3.0")
	_ = st.SetInfoBlock(2, []byte("hello"))
	_ = st.Close()
	a, out, _ := newTestApp(ctl)
	if err := a.ReadInfoBlock(all, 2, false); err != nil {
		t.Fatal(err)
	}
	if want := "desk: hello  (68 65 6c 6c 6f)\nBS073788-3.1: -\n"; out.String() != want {
		t.Errorf("got %q", out.String())
	}
}

func TestReadInfoBlockJSON(t *testing.T) {
	ctl := deskAndSquare()
	st, _ := ctl.Open("BS072777-3.0")
	_ = st.SetInfoBlock(2, []byte("hi"))
	_ = st.Close()
	a, out, _ := newTestApp(ctl)
	if err := a.ReadInfoBlock(dev("desk"), 2, true); err != nil {
		t.Fatal(err)
	}
	var got []struct {
		Block     int `json:"block"`
		InfoBlock struct {
			Text string `json:"text"`
			Hex  string `json:"hex"`
		} `json:"info_block"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Block != 2 || got[0].InfoBlock.Text != "hi" || got[0].InfoBlock.Hex != "68 69" {
		t.Errorf("got %+v", got)
	}
}

func TestSetInfoBlock(t *testing.T) {
	ctl := deskAndSquare()
	a, _, errb := newTestApp(ctl)
	if err := a.SetInfoBlock(dev("desk"), 2, "hello", false); err != nil {
		t.Fatal(err)
	}
	if got := string(ctl.Stick("BS072777-3.0").Stored(2)); got != "hello" || errb.Len() != 0 {
		t.Errorf("stored %q stderr %q", got, errb.String())
	}
	if err := a.SetInfoBlock(dev("desk"), 2, "68 69", true); err != nil {
		t.Fatal(err)
	}
	if got := string(ctl.Stick("BS072777-3.0").Stored(2)); got != "hi" {
		t.Errorf("stored %q", got)
	}
}

func TestSetInfoBlockOneWarnsAboutTheName(t *testing.T) {
	a, _, errb := newTestApp(deskAndSquare())
	if err := a.SetInfoBlock(dev("desk"), 1, "x", false); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errb.String(), "holds the stick's name") {
		t.Errorf("stderr %q", errb.String())
	}
}

func TestSetInfoBlockRefuses(t *testing.T) {
	cases := []struct {
		name  string
		block int
		data  string
		isHex bool
		all   bool
	}{
		{"block 3", 3, "x", false, false},
		{"bad hex", 2, "zz", true, false},
		{"too long", 2, strings.Repeat("x", 33), false, false},
		{"group", 2, "x", false, true},
	}
	for _, tc := range cases {
		a, _, _ := newTestApp(deskAndSquare())
		req := dev("desk")
		if tc.all {
			req = all
		}
		if err := a.SetInfoBlock(req, tc.block, tc.data, tc.isHex); exitcode.From(err) != 2 {
			t.Errorf("%s: err %v", tc.name, err)
		}
	}
}
