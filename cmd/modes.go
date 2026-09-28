package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/bawdo/blinky/internal/app"
	"github.com/bawdo/blinky/internal/colour"
	"github.com/bawdo/blinky/internal/target"
)

func newDiscoCmd(a *app.App) *cobra.Command {
	var o app.DiscoOptions
	c := &cobra.Command{
		Use:               "disco [<colour>...]",
		Short:             "Pulse every LED in random colours until stopped",
		Args:              validArgs(cobra.ArbitraryArgs),
		ValidArgsFunction: colourCompleter(0),
	}
	c.Long = long(c.Short, target.Help(target.Group, target.Group),
		"Each LED picks a colour, fades it up and down over a random period, then stays dark for a random gap. "+
			"With colours given, it picks from those, otherwise from bright random colours. "+colourHelp,
		stopHelp)
	t := addTargetFlags(c, a)
	l := addLEDFlags(c)
	durationFlag(c, &o.MinPeriod, "min-period", 200*time.Millisecond, "shortest pulse")
	durationFlag(c, &o.MaxPeriod, "max-period", 2*time.Second, "longest pulse")
	durationFlag(c, &o.MaxGap, "max-gap", time.Second, "longest dark gap between pulses")
	durationFlag(c, &o.Duration, "duration", 0, "stop after this long, 0 for until stopped")
	c.Flags().Uint64Var(&o.Seed, "seed", 0, "repeat a run exactly, 0 for a random seed")
	_ = c.Flags().MarkHidden("seed")
	c.RunE = func(cmd *cobra.Command, args []string) error {
		set, err := l.settings()
		if err != nil {
			return err
		}
		if o.Palette, err = colour.ParseAll(args); err != nil {
			return err
		}
		return a.Disco(cmd.Context(), t.request(), set, o)
	}
	return c
}

func newPoliceCmd(a *app.App) *cobra.Command {
	var o app.PoliceOptions
	c := &cobra.Command{
		Use:   "police",
		Short: "Crossfade red and blue halves until stopped",
		Args:  validArgs(cobra.NoArgs),
	}
	c.Long = long(c.Short, target.Help(target.Group, target.Group),
		"Half of each stick's LEDs fade from off to red while the other half fade from blue to off, then back. "+
			"Every stick runs in step; --alternate puts neighbouring sticks in opposite phase.",
		stopHelp)
	t := addTargetFlags(c, a)
	l := addLEDFlags(c)
	durationFlag(c, &o.Period, "period", time.Second, "one full red and blue swap")
	c.Flags().BoolVar(&o.Alternate, "alternate", false, "every second stick runs in opposite phase")
	durationFlag(c, &o.Duration, "duration", 0, "stop after this long, 0 for until stopped")
	c.RunE = func(cmd *cobra.Command, _ []string) error {
		set, err := l.settings()
		if err != nil {
			return err
		}
		return a.Police(cmd.Context(), t.request(), set, o)
	}
	return c
}
