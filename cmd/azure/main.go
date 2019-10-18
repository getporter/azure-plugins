package main

import (
	"os"

	"github.com/deislabs/porter-azure-plugins/pkg/azure"
	"github.com/spf13/cobra"
)

func main() {
	cmd := buildRootCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func buildRootCommand() *cobra.Command {
	p := azure.New()

	cmd := &cobra.Command{
		Use:   "azure",
		Short: "Azure plugins for Porter",
	}

	cmd.AddCommand(buildVersionCommand(p))
	cmd.AddCommand(buildRunCommand(p))

	return cmd
}
