package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"get.porter.sh/plugin/azure/pkg/azure"
	"get.porter.sh/porter/pkg/cli"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	run := func() int {
		ctx := context.Background()
		m := azure.New()
		ctx, err := m.ConfigureLogging(ctx)
		if err != nil {
			fmt.Println(err)
			os.Exit(cli.ExitCodeErr)
		}
		cmd := buildRootCommand(m, getInput())

		ctx, log := m.StartRootSpan(ctx, "azure")
		defer func() {
			// Capture panics and trace them
			if panicErr := recover(); panicErr != nil {
				log.Error(fmt.Errorf("%s", panicErr),
					attribute.Bool("panic", true),
					attribute.String("stackTrace", string(debug.Stack())))
				log.EndSpan()
				m.Close()
				os.Exit(cli.ExitCodeErr)
			} else {
				log.Close()
				m.Close()
			}
		}()

		if err := cmd.ExecuteContext(ctx); err != nil {
			return cli.ExitCodeErr
		}
		return cli.ExitCodeSuccess
	}
	os.Exit(run())
}

func buildRootCommand(m *azure.Plugin, in io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "azure",
		Short: "Azure plugin for Porter",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Enable swapping out stdout/stderr for testing
			m.In = in
			m.Out = cmd.OutOrStdout()
			m.Err = cmd.OutOrStderr()
		},
	}

	cmd.AddCommand(buildVersionCommand(m))
	cmd.AddCommand(buildRunCommand(m))

	return cmd
}
func getInput() io.Reader {
	s, _ := os.Stdin.Stat()
	if (s.Mode() & os.ModeCharDevice) == 0 {
		return os.Stdin
	}

	return &bytes.Buffer{}
}
