package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/bawdo/blinky/internal/app"
	"github.com/bawdo/blinky/internal/colour"
	"github.com/bawdo/blinky/internal/settings"
	"github.com/bawdo/blinky/internal/target"
)

type repeatFunc func(context.Context, target.Request, settings.Settings, colour.Spec, app.Repeat) error

func newBlinkCmd(a *app.App) *cobra.Command {
	var (
		o       app.BlinkOptions
		on, off time.Duration
		second  string
	)
	c := &cobra.Command{
		Use:               "blink <colour>",
		Short:             "Blink LEDs on and off",
		Args:              validArgs(cobra.ExactArgs(1)),
		ValidArgsFunction: colourCompleter(1),
	}
	c.Long = long(c.Short, target.Help(target.Group, target.Group), colourHelp,
		"--on-time and --off-time each default to half of --period. With --second-colour, "+
			"each repeat blinks the colour, then the second colour.",
		stopHelp)
	t := addTargetFlags(c, a)
	l := addLEDFlags(c)
	durationFlag(c, &o.Period, "period", time.Second, "one on and off cycle")
	durationFlag(c, &on, "on-time", 0, "how long the LEDs stay lit, default half of --period")
	durationFlag(c, &off, "off-time", 0, "how long the LEDs stay dark, default half of --period")
	colourFlag(c, &second, "second-colour", "alternate with this colour")
	c.Flags().IntVar(&o.Repeats, "repeats", 3, "how many times, 0 for until stopped")
	durationFlag(c, &o.Duration, "duration", 0, "stop after this long, 0 for no limit")
	c.RunE = func(cmd *cobra.Command, args []string) error {
		set, err := l.settings()
		if err != nil {
			return err
		}
		spec, err := colour.Parse(args[0])
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("on-time") {
			o.On = &on
		}
		if cmd.Flags().Changed("off-time") {
			o.Off = &off
		}
		if o.Second, err = optionalColour(cmd, "second-colour", second); err != nil {
			return err
		}
		return a.Blink(cmd.Context(), t.request(), set, spec, o)
	}
	return c
}

func newPulseCmd(a *app.App) *cobra.Command {
	return newRepeatCmd(a, "pulse", "Fade LEDs up and down", 2*time.Second, "one fade up and down", a.Pulse)
}

// newRepeatCmd builds pulse.
func newRepeatCmd(a *app.App, name, short string, period time.Duration, periodUsage string, fn repeatFunc) *cobra.Command {
	var r app.Repeat
	c := &cobra.Command{
		Use:               name + " <colour>",
		Short:             short,
		Args:              validArgs(cobra.ExactArgs(1)),
		ValidArgsFunction: colourCompleter(1),
	}
	c.Long = long(c.Short, target.Help(target.Group, target.Group), colourHelp, stopHelp)
	t := addTargetFlags(c, a)
	l := addLEDFlags(c)
	durationFlag(c, &r.Period, "period", period, periodUsage)
	c.Flags().IntVar(&r.Repeats, "repeats", 3, "how many times, 0 for until stopped")
	durationFlag(c, &r.Duration, "duration", 0, "stop after this long, 0 for no limit")
	c.RunE = func(cmd *cobra.Command, args []string) error {
		set, err := l.settings()
		if err != nil {
			return err
		}
		spec, err := colour.Parse(args[0])
		if err != nil {
			return err
		}
		return fn(cmd.Context(), t.request(), set, spec, r)
	}
	return c
}

func newMorphCmd(a *app.App) *cobra.Command {
	var d time.Duration
	c := &cobra.Command{
		Use:               "morph <colour>",
		Short:             "Fade LEDs from their current colour to a new one",
		Args:              validArgs(cobra.ExactArgs(1)),
		ValidArgsFunction: colourCompleter(1),
	}
	c.Long = long(c.Short, target.Help(target.Group, target.Group), colourHelp,
		"Ctrl-C stops it and turns the LEDs off. Left to finish, the LEDs stay on the new colour.")
	t := addTargetFlags(c, a)
	l := addLEDFlags(c)
	durationFlag(c, &d, "duration", time.Second, "how long the fade takes")
	c.RunE = func(cmd *cobra.Command, args []string) error {
		set, err := l.settings()
		if err != nil {
			return err
		}
		spec, err := colour.Parse(args[0])
		if err != nil {
			return err
		}
		return a.Morph(cmd.Context(), t.request(), set, spec, app.MorphOptions{Fade: d})
	}
	return c
}
