package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bawdo/blinky/internal/app"
)

func newListCmd(a *app.App) *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "list",
		Short: "List attached BlinkSticks and the ID to use with --device",
		Args:  validArgs(cobra.NoArgs),
		RunE: func(*cobra.Command, []string) error {
			return a.List(asJSON)
		},
	}
	addJSONFlag(c, &asJSON)
	return c
}
