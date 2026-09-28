package stick

import (
	"slices"
	"strings"

	"github.com/bawdo/go-blinkstick"
)

// Hardware is the Controller for sticks attached to this machine.
type Hardware struct{}

// List reads every attached stick's name, so it briefly opens each one.
func (Hardware) List() ([]Info, error) {
	named, err := blinkstick.ListNamed()
	if err != nil {
		return nil, err
	}
	infos := make([]Info, len(named))
	for i, n := range named {
		infos[i] = fromNamed(n)
	}
	SortBySerial(infos)
	return infos, nil
}

// Open opens the stick with the given serial.
func (Hardware) Open(serial string) (Stick, error) {
	d, err := blinkstick.OpenSerial(serial)
	if err != nil {
		return nil, err
	}
	return device{d}, nil
}

// SortBySerial sorts infos in place by serial, the order list shows and
// police --alternate uses.
func SortBySerial(infos []Info) {
	slices.SortFunc(infos, func(a, b Info) int { return strings.Compare(a.Serial, b.Serial) })
}

func fromNamed(n blinkstick.NamedInfo) Info {
	status := StatusOK
	switch {
	case n.Busy:
		status = StatusBusy
	case n.Model.Name == "unknown":
		status = StatusUnsupported
	}
	return Info{
		Serial:       n.Serial,
		Firmware:     n.Version,
		Manufacturer: n.Manufacturer,
		Product:      n.Product,
		Model:        n.Model.Name,
		LEDs:         n.Model.LEDs,
		Name:         n.Name,
		Status:       status,
	}
}

// device adapts *blinkstick.Device to Stick.
type device struct{ *blinkstick.Device }

func (d device) LEDs() int { return d.Info().Model.LEDs }

var _ Stick = device{}
