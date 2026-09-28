package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/bawdo/blinky/internal/app"
	"github.com/bawdo/blinky/internal/colour"
	"github.com/bawdo/blinky/internal/stick"
)

// completeDevices offers the IDs of attached sticks, described by model
// and serial.
func completeDevices(a *app.App) cobra.CompletionFunc {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		infos, err := a.Sticks()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		var out []cobra.Completion
		for _, i := range infos {
			if i.Status != stick.StatusUnsupported && strings.HasPrefix(i.ID(), toComplete) {
				out = append(out, cobra.CompletionWithDesc(i.ID(), i.Model+" "+i.Serial))
			}
		}
		return out, cobra.ShellCompDirectiveNoFileComp
	}
}

// colourCompleter offers colour keywords and CSS names for up to maxArgs
// positional arguments; 0 means any number.
func colourCompleter(maxArgs int) cobra.CompletionFunc {
	return func(_ *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if maxArgs > 0 && len(args) >= maxArgs {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		prefix := strings.ToLower(toComplete)
		var out []cobra.Completion
		for _, s := range colour.Suggestions() {
			if strings.HasPrefix(s.Value, prefix) {
				out = append(out, cobra.CompletionWithDesc(s.Value, s.Description))
			}
		}
		return out, cobra.ShellCompDirectiveNoFileComp
	}
}

func completeDurations(_ *cobra.Command, _ []string, _ string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return []cobra.Completion{"500ms", "1s", "5s", "30s"}, cobra.ShellCompDirectiveNoFileComp
}
