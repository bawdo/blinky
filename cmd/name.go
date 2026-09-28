package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bawdo/blinky/internal/app"
	"github.com/bawdo/blinky/internal/exitcode"
	"github.com/bawdo/blinky/internal/target"
)

func newNameCmd(a *app.App) *cobra.Command {
	var clearName, asJSON bool
	c := &cobra.Command{
		Use:               "name [<name>]",
		Short:             "Read or set the name stored on a stick",
		Args:              validArgs(cobra.MaximumNArgs(1)),
		ValidArgsFunction: cobra.NoFileCompletions,
	}
	c.Long = long(c.Short, target.Help(target.Group, target.One),
		"Names are UTF-8, at most 32 bytes, unique among attached sticks, and must not look like a serial. "+
			"Once set, use the name with --device.",
		eepromHelp)
	t := addTargetFlags(c, a)
	c.Flags().BoolVar(&clearName, "clear", false, "remove the stored name")
	addJSONFlag(c, &asJSON)
	c.RunE = func(_ *cobra.Command, args []string) error {
		switch {
		case clearName && len(args) > 0:
			return exitcode.Invalid("give a name or --clear, not both")
		case asJSON && (clearName || len(args) > 0):
			return exitcode.Invalid("--json only applies when reading")
		case len(args) == 1 && args[0] == "":
			return exitcode.Invalid("name is empty, use --clear to remove it")
		case clearName:
			return a.SetName(t.request(), "")
		case len(args) == 1:
			return a.SetName(t.request(), args[0])
		}
		return a.ReadName(t.request(), asJSON)
	}
	return c
}
