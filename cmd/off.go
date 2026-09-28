package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bawdo/blinky/internal/app"
	"github.com/bawdo/blinky/internal/target"
)

func newOffCmd(a *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "off",
		Short: "Turn LEDs off",
		Args:  validArgs(cobra.NoArgs),
	}
	c.Long = long(c.Short, target.Help(target.Group, target.Group))
	t := addTargetFlags(c, a)
	l := addLEDFlags(c)
	c.RunE = func(*cobra.Command, []string) error {
		set, err := l.settings()
		if err != nil {
			return err
		}
		return a.Off(t.request(), set)
	}
	return c
}
