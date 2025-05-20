package commands

import (
	"context"
	// "fmt"
	// "io"
	// "fmt"
	// "log/slog"
	// "os"

	"github.com/ainvaltin/nu-plugin"
	"github.com/ainvaltin/nu-plugin/types"
	// "github.com/gtnebel/nu_plugin_nuplot/commands/flags"
)

func Nuplot() *nu.Command {
	return &nu.Command{
		Signature: nu.PluginSignature{
			Name:        "nuplot",
			Category:    "Chart",
			Desc:        "nuplot is a nushell plugin for plotting charts. It builds interactive charts from your data that are opened inside the web browser.",
			Description: "You must use one of the following subcommands. Using this command as-is will only produce this help message.",
			SearchTerms: []string{"plot", "graph"},
			// OptionalPositional: nu.PositionalArgs{},
			Named: nu.Flags{},
			InputOutputTypes: []nu.InOutTypes{
				{In: types.Nothing(), Out: types.Nothing()},
			},
			AllowMissingExamples: true,
		},
		Examples: nu.Examples{},
		OnRun:    nuplotHandler,
	}
}

func nuplotHandler(ctx context.Context, call *nu.ExecCommand) error {
	h, _ := call.GetHelp(ctx)
	println(h)

	return nil
}
