package azure

import (
	"github.com/deislabs/porter-azure-plugins/pkg/azure/blob"
	"github.com/deislabs/porter/pkg/instance-storage/claimstore"
	"github.com/deislabs/porter/pkg/plugins"
	"github.com/hashicorp/go-plugin"
)

func Run() {
	p := map[string]plugin.Plugin{
		claimstore.PluginKey: blob.NewPlugin(),
	}
	plugins.ServeMany(p)
}
