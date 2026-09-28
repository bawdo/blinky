package cmd

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/bawdo/blinky/internal/app"
	"github.com/bawdo/blinky/internal/exitcode"
	"github.com/bawdo/blinky/internal/target"
)

func newInfoBlockCmd(a *app.App) *cobra.Command {
	var clearBlock, isHex, asJSON bool
	c := &cobra.Command{
		Use:   "info-block <1|2> [<data>]",
		Short: "Read or write a stick's 32 byte info blocks",
		Args:  validArgs(cobra.RangeArgs(1, 2)),
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]cobra.Completion, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []cobra.Completion{"1", "2"}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
	c.Long = long(c.Short, target.Help(target.Group, target.One),
		"Block 1 holds the stick's name. Data is text, or bytes with --hex (68656c6c6f or \"68 65 6c 6c 6f\").",
		eepromHelp)
	t := addTargetFlags(c, a)
	c.Flags().BoolVar(&clearBlock, "clear", false, "empty the block")
	c.Flags().BoolVar(&isHex, "hex", false, "data is hex bytes")
	addJSONFlag(c, &asJSON)
	c.RunE = func(_ *cobra.Command, args []string) error {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return exitcode.Invalid("info block %q, want 1 or 2", args[0])
		}
		writing := clearBlock || len(args) == 2
		switch {
		case clearBlock && len(args) == 2:
			return exitcode.Invalid("give data or --clear, not both")
		case asJSON && writing:
			return exitcode.Invalid("--json only applies when reading")
		case isHex && !writing:
			return exitcode.Invalid("--hex only applies when writing")
		case clearBlock:
			return a.SetInfoBlock(t.request(), n, "", false)
		case len(args) == 2:
			return a.SetInfoBlock(t.request(), n, args[1], isHex)
		}
		return a.ReadInfoBlock(t.request(), n, asJSON)
	}
	return c
}
