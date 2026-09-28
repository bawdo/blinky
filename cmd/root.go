package cmd

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/bawdo/blinky/internal/app"
	"github.com/bawdo/blinky/internal/exitcode"
)

// Execute runs the CLI. The returned int is the process exit code.
func Execute() int {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	// After the first signal, restore the default handler so a second
	// Ctrl-C quits at once even if a stick is not responding.
	go func() {
		<-ctx.Done()
		stop()
	}()
	err := newRootCmd(app.Options{}).ExecuteContext(ctx)
	return classifyError(err)
}

// classifyError maps an error to documented exit codes. Delegates to
// exitcode.From so the sentinel-to-code map has a single source of truth.
// Document new codes in README.md when you add sentinels.
func classifyError(err error) int {
	return exitcode.From(err)
}

func newRootCmd(opts app.Options) *cobra.Command {
	a := app.NewWithOptions(opts)
	root := &cobra.Command{
		Use:           "blinky",
		Short:         "Control BlinkStick Nano and Square LEDs",
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return fmt.Errorf("%w: %w", exitcode.ErrInvalidArgs, err)
	})
	root.AddCommand(
		newVersionCmd(),
		newListCmd(a),
		newInfoCmd(a),
		newColourCmd(a, "colour", false),
		newColourCmd(a, "color", true),
		newOffCmd(a),
		newNameCmd(a),
		newInfoBlockCmd(a),
		newBlinkCmd(a),
		newPulseCmd(a),
		newMorphCmd(a),
		newDiscoCmd(a),
		newPoliceCmd(a),
	)
	applyHelpBareword(root)
	return root
}

func applyHelpBareword(cmd *cobra.Command) {
	orig := cmd.RunE
	if orig != nil {
		cmd.RunE = func(c *cobra.Command, args []string) error {
			if len(args) > 0 && args[0] == "help" {
				return c.Help()
			}
			return orig(c, args)
		}
	}
	for _, sub := range cmd.Commands() {
		applyHelpBareword(sub)
	}
}
