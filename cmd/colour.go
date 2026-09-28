package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bawdo/blinky/internal/app"
	"github.com/bawdo/blinky/internal/colour"
	"github.com/bawdo/blinky/internal/exitcode"
	"github.com/bawdo/blinky/internal/target"
)

// newColourCmd builds colour, and its hidden copy color.
func newColourCmd(a *app.App, name string, hidden bool) *cobra.Command {
	var led int
	var asJSON bool
	c := &cobra.Command{
		Use:               name + " [<colour>...]",
		Short:             "Read or set LED colours",
		Hidden:            hidden,
		Args:              validArgs(cobra.ArbitraryArgs),
		ValidArgsFunction: colourCompleter(0),
	}
	c.Long = long(c.Short, target.Help(target.Group, target.Group),
		"With no colours it prints each stick's LEDs. "+colourHelp,
		"Given N colours, each stick splits its LEDs into N even groups: with red blue a Nano shows "+
			"red then blue, and a Square shows four red then four blue.")
	t := addTargetFlags(c, a)
	l := addLEDFlags(c)
	c.Flags().IntVar(&led, "led", -1, "one physical LED, counting from 0")
	addJSONFlag(c, &asJSON)
	c.RunE = func(cmd *cobra.Command, args []string) error {
		set, err := l.settings()
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("led") && led < 0 {
			return exitcode.Invalid("--led %d, LEDs count from 0", led)
		}
		if len(args) == 0 {
			return a.ReadColour(t.request(), set, led, asJSON)
		}
		if asJSON {
			return exitcode.Invalid("--json only applies when reading")
		}
		specs, err := colour.ParseAll(args)
		if err != nil {
			return err
		}
		return a.SetColour(t.request(), set, led, specs)
	}
	return c
}
