package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/bawdo/blinky/internal/app"
	"github.com/bawdo/blinky/internal/exitcode"
	"github.com/bawdo/blinky/internal/settings"
	"github.com/bawdo/blinky/internal/target"
)

// Help text shared by several commands.
const (
	colourHelp = "Colours can be hex (ff8800, f80, or quoted '#ff8800'), r,g,b (255,136,0), " +
		"one of the 148 CSS names (cornflowerblue), or off, random or vivid."
	stopHelp   = "Ctrl-C, or --duration running out, stops it and turns the LEDs off."
	eepromHelp = "Writing stores data in EEPROM on the stick, which wears out with heavy use. " +
		"Do not run it in a loop."
)

// long builds a command's long help from its short description and extra
// paragraphs, such as its "Targets:" line.
func long(short string, parts ...string) string {
	return short + ".\n\n" + strings.Join(parts, "\n\n")
}

// validArgs wraps a Cobra argument validator so its errors exit 2.
func validArgs(v cobra.PositionalArgs) cobra.PositionalArgs {
	return func(c *cobra.Command, args []string) error {
		if err := v(c, args); err != nil {
			return fmt.Errorf("%w: %w", exitcode.ErrInvalidArgs, err)
		}
		return nil
	}
}

type targetFlags struct {
	devices []string
	all     bool
}

// addTargetFlags adds --device and --all to c.
func addTargetFlags(c *cobra.Command, a *app.App) *targetFlags {
	f := &targetFlags{}
	c.Flags().StringArrayVarP(&f.devices, "device", "d", nil, "stick serial or name, repeatable (see blinky list)")
	c.Flags().BoolVarP(&f.all, "all", "a", false, "every attached stick")
	_ = c.RegisterFlagCompletionFunc("device", completeDevices(a))
	return f
}

func (f *targetFlags) request() target.Request {
	return target.Request{Devices: f.devices, All: f.all}
}

type ledFlags struct {
	c          *cobra.Command
	brightness int
	inverse    bool
}

// addLEDFlags adds --brightness and --inverse to a command that writes
// LEDs.
func addLEDFlags(c *cobra.Command) *ledFlags {
	f := &ledFlags{c: c}
	c.Flags().IntVar(&f.brightness, "brightness", 100, "brightness limit in percent, 0 to 100")
	c.Flags().BoolVar(&f.inverse, "inverse", false, "flip every colour, for LEDs wired so 255 means off")
	return f
}

// settings resolves the flags layer. Only flags the user set count, so a
// config file layer can sit beneath them later.
func (f *ledFlags) settings() (settings.Settings, error) {
	var l settings.Layer
	if f.c.Flags().Changed("brightness") {
		l.Brightness = &f.brightness
	}
	if f.c.Flags().Changed("inverse") {
		l.Inverse = &f.inverse
	}
	return settings.Resolve(l)
}

func addJSONFlag(c *cobra.Command, p *bool) {
	c.Flags().BoolVar(p, "json", false, "print JSON")
}

// durationFlag adds a duration flag with completion of common values.
func durationFlag(c *cobra.Command, p *time.Duration, name string, def time.Duration, usage string) {
	c.Flags().DurationVar(p, name, def, usage)
	_ = c.RegisterFlagCompletionFunc(name, completeDurations)
}
