package sticktest

import (
	"errors"
	"testing"

	"github.com/bawdo/go-blinkstick"
)

func TestFakeScalesByLimit(t *testing.T) {
	c := New(Nano("BS1", ""))
	st, err := c.Open("BS1")
	if err != nil {
		t.Fatal(err)
	}
	st.SetBrightnessLimit(128)
	if err := st.SetFrame([]blinkstick.RGB{{R: 255}, {B: 255}}); err != nil {
		t.Fatal(err)
	}
	if got := c.Stick("BS1").Current(); got[0] != (blinkstick.RGB{R: 128}) || got[1] != (blinkstick.RGB{B: 128}) {
		t.Errorf("stored %v", got)
	}
}

func TestFakeOpenUnknownSerial(t *testing.T) {
	if _, err := New().Open("BS9"); !errors.Is(err, blinkstick.ErrNotFound) {
		t.Errorf("error = %v, want ErrNotFound", err)
	}
}

func TestFakeListReflectsNewName(t *testing.T) {
	c := New(Nano("BS1", ""))
	st, _ := c.Open("BS1")
	if err := st.SetName("desk"); err != nil {
		t.Fatal(err)
	}
	infos, _ := c.List()
	if infos[0].Name != "desk" {
		t.Errorf("List name = %q", infos[0].Name)
	}
}

func TestFakeRejectsUseAfterClose(t *testing.T) {
	c := New(Nano("BS1", ""))
	st, _ := c.Open("BS1")
	_ = st.Close()
	if err := st.Off(); !errors.Is(err, blinkstick.ErrClosed) {
		t.Errorf("error = %v, want ErrClosed", err)
	}
	if !c.Stick("BS1").Closed() {
		t.Error("Closed() = false")
	}
}
