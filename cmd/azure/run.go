package main

import (
	"github.com/deislabs/porter-azure-plugins/pkg/azure"
	"github.com/spf13/cobra"
)

func buildRunCommand(p *azure.Plugin) *cobra.Command {
	var opts azure.RunOptions

	cmd := &cobra.Command{
		Use:   "run PLUGIN_IMPLEMENTATION",
		Short: "Run the plugin and listen for client connections.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		Run: func(cmd *cobra.Command, args []string) {
			p.Run(opts)
		},
	}

	return cmd
}
