package azure

import (
	"github.com/deislabs/porter-azure-plugins/pkg/azure/blob"
	"github.com/deislabs/porter/pkg/instance-storage/claimstore"
	"github.com/deislabs/porter/pkg/plugins"
	"github.com/hashicorp/go-plugin"
)

func Run() {
	// TODO: decide which implementation to use based on the argument passed to the plugin binary

	p := map[string]plugin.Plugin{
		claimstore.PluginKey: blob.NewPlugin(),
	}
	plugins.ServeMany(p)
}
