package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/ainvaltin/nu-plugin"
	"github.com/ainvaltin/nu-plugin/types"
)

// This function initializes the nuplot main command. This command prints
// the help message of the nuplot plugin to stderr.
func Nuplot() *nu.Command {
	return &nu.Command{
		Signature: nu.PluginSignature{
			Name:        "nuplot",
			Category:    "Chart",
			Desc:        "nuplot is a nushell plugin for plotting charts. It builds interactive charts from your data that are opened inside the web browser.",
			Description: "You must use one of the following subcommands. Using this command as-is will only produce this help message.",
			SearchTerms: []string{"plot", "graph"},
			Named:       []nu.Flag{},
			InputOutputTypes: []nu.InOutTypes{
				{In: types.Nothing(), Out: types.Nothing()},
			},
			AllowMissingExamples: true,
		},
		Examples: []nu.Example{},
		OnRun:    nuplotHandler,
	}
}

func nuplotHandler(ctx context.Context, call *nu.ExecCommand) error {
	h, _ := call.GetHelp(ctx)
	fmt.Fprintln(os.Stderr, h)

	return nil
}
