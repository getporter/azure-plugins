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
	cmd := &cobra.Command{
		Use: "azure",
		Short: "Azure plugins for Porter",
		Run: func(cmd *cobra.Command, args []string) {
			azure.Run()
		},
	}
	return cmd
}