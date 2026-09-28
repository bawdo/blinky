package app

import (
	"fmt"
	"strconv"

	"github.com/bawdo/blinky/internal/render"
	"github.com/bawdo/blinky/internal/stick"
	"github.com/bawdo/blinky/internal/target"
)

type stickJSON struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Serial       string `json:"serial"`
	Model        string `json:"model"`
	LEDs         int    `json:"leds"`
	Firmware     string `json:"firmware"`
	Manufacturer string `json:"manufacturer"`
	Product      string `json:"product"`
	Status       string `json:"status"`
}

func sticksJSON(infos []stick.Info) []stickJSON {
	out := make([]stickJSON, 0, len(infos))
	for _, i := range infos {
		out = append(out, stickJSON{ID: i.ID(), Name: i.Name, Serial: i.Serial, Model: i.Model, LEDs: i.LEDs,
			Firmware: i.Firmware, Manufacturer: i.Manufacturer, Product: i.Product, Status: string(i.Status)})
	}
	return out
}

// List prints every attached stick, with the ID to pass to --device first.
func (a *App) List(asJSON bool) error {
	infos, err := a.ctl.List()
	if err != nil {
		return fmt.Errorf("listing BlinkSticks: %w", err)
	}
	if asJSON {
		return render.JSON(a.out, sticksJSON(infos))
	}
	if len(infos) == 0 {
		_, _ = fmt.Fprintln(a.err, "No BlinkSticks attached.")
		return nil
	}
	rows := make([][]string, len(infos))
	for k, i := range infos {
		rows[k] = []string{i.ID(), nameCell(i), i.Serial, i.Model, ledsCell(i), string(i.Status)}
	}
	return render.Table(a.out, []string{"ID", "NAME", "SERIAL", "MODEL", "LEDS", "STATUS"}, rows)
}

// Info prints details of the chosen sticks. It reads the stick list only,
// so it works on busy sticks too.
func (a *App) Info(req target.Request, asJSON bool) error {
	infos, _, err := a.resolve(req, target.Group)
	if err != nil {
		return err
	}
	if asJSON {
		return render.JSON(a.out, sticksJSON(infos))
	}
	for k, i := range infos {
		if k > 0 {
			if _, err := fmt.Fprintln(a.out); err != nil {
				return err
			}
		}
		err := render.Fields(a.out, i.ID(), [][2]string{
			{"Serial", i.Serial}, {"Model", modelCell(i)}, {"Firmware", i.Firmware},
			{"Manufacturer", i.Manufacturer}, {"Product", i.Product}, {"Name", nameCell(i)},
			{"Status", string(i.Status)},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Sticks lists the attached sticks, for shell completion.
func (a *App) Sticks() ([]stick.Info, error) {
	return a.ctl.List()
}

// nameCell shows "?" when a busy stick's name cannot be read and "-" when
// there is no name.
func nameCell(i stick.Info) string {
	switch {
	case i.Status == stick.StatusBusy:
		return "?"
	case i.Name == "":
		return "-"
	}
	return i.Name
}

func ledsCell(i stick.Info) string {
	if i.Status == stick.StatusUnsupported {
		return "-"
	}
	return strconv.Itoa(i.LEDs)
}

func modelCell(i stick.Info) string {
	if i.Status == stick.StatusUnsupported {
		return i.Model
	}
	return fmt.Sprintf("%s (%d LEDs)", i.Model, i.LEDs)
}
