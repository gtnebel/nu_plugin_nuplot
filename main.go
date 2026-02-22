// nu_plugin_nuplot is a [Nushell](https://nushell.sh/) plugin for plotting charts.
// It builds interactive charts from your data that are opened inside the web
// browser.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ainvaltin/nu-plugin"

	"github.com/gtnebel/nu_plugin_nuplot/commands"
)

// The plugin version that is passed to nushell when the plugin is registered.
const PluginVersion = "0.2.4"

// global system signal handler
func quitSignalContext() context.Context {
	ctx, cancel := context.WithCancelCause(context.Background())

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(sigChan)
		sig := <-sigChan
		cancel(fmt.Errorf("got quit signal: %s", sig))
	}()

	return ctx
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelInfo)

	p, err := nu.New(
		[]*nu.Command{
			commands.Nuplot(),
			commands.NuplotLine(),
			commands.NuplotKline(),
			commands.NuplotBar(),
			commands.NuplotPie(),
			commands.NuplotBoxPlot(),
		},
		PluginVersion,
		nil,
	)
	if err != nil {
		slog.Error("failed to create plugin", "error", err)
		return
	}
	if err := p.Run(quitSignalContext()); err != nil && !errors.Is(err, nu.ErrGoodbye) {
		slog.Error("plugin exited with error", "error", err)
	}
}
