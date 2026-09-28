package cmd

import (
	"slices"
	"testing"
)

func TestNameCommand(t *testing.T) {
	ctl := twoSticks()
	if res := run(t, opts(ctl), "name", "-d", "BS073788-3.1", "shelf"); res.code != 0 {
		t.Fatalf("exit %d err %q", res.code, res.err)
	}
	if res := run(t, opts(ctl), "name", "-a"); res.out != "desk: desk\nshelf: shelf\n" {
		t.Errorf("read %q", res.out)
	}
	if res := run(t, opts(ctl), "name", "-d", "shelf", "--clear"); res.code != 0 {
		t.Fatalf("clear: exit %d err %q", res.code, res.err)
	}
	if got := ctl.Stick("BS073788-3.1").StoredName(); got != "" {
		t.Errorf("stored %q", got)
	}
}

func TestNameCommandRejects(t *testing.T) {
	cases := [][]string{
		{"name", "-d", "desk", "x", "--clear"},
		{"name", "-d", "desk", "x", "--json"},
		{"name", "-d", "desk", "a", "b"},
		{"name", "-a", "x"},
	}
	for _, args := range cases {
		if res := run(t, opts(twoSticks()), args...); res.code != 2 {
			t.Errorf("%v: exit %d, want 2", args, res.code)
		}
	}
}

func TestInfoBlockCommand(t *testing.T) {
	ctl := twoSticks()
	if res := run(t, opts(ctl), "info-block", "-d", "desk", "2", "--hex", "6869"); res.code != 0 {
		t.Fatalf("exit %d err %q", res.code, res.err)
	}
	if res := run(t, opts(ctl), "info-block", "-d", "desk", "2"); res.out != "desk: hi  (68 69)\n" {
		t.Errorf("read %q", res.out)
	}
	if res := run(t, opts(ctl), "info-block", "-d", "desk", "2", "--clear"); res.code != 0 {
		t.Fatalf("clear: exit %d", res.code)
	}
	if got := ctl.Stick("BS072777-3.0").Stored(2); len(got) != 0 {
		t.Errorf("stored %q", got)
	}
}

func TestInfoBlockCommandRejects(t *testing.T) {
	cases := [][]string{
		{"info-block", "-d", "desk", "3"},
		{"info-block", "-d", "desk", "two"},
		{"info-block", "-d", "desk", "2", "x", "--clear"},
		{"info-block", "-d", "desk", "2", "--hex"},
		{"info-block", "-d", "desk"},
	}
	for _, args := range cases {
		if res := run(t, opts(twoSticks()), args...); res.code != 2 {
			t.Errorf("%v: exit %d, want 2", args, res.code)
		}
	}
}

func TestInfoBlockCompletion(t *testing.T) {
	if got := complete(t, opts(twoSticks()), "info-block", ""); !slices.Equal(got, []string{"1", "2"}) {
		t.Errorf("got %v", got)
	}
}
