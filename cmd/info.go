package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bawdo/blinky/internal/app"
	"github.com/bawdo/blinky/internal/target"
)

func newInfoCmd(a *app.App) *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "info",
		Short: "Show serial, model, firmware and name of sticks",
		Args:  validArgs(cobra.NoArgs),
	}
	c.Long = long(c.Short, target.Help(target.Group, target.Group))
	t := addTargetFlags(c, a)
	addJSONFlag(c, &asJSON)
	c.RunE = func(*cobra.Command, []string) error {
		return a.Info(t.request(), asJSON)
	}
	return c
}
