package blob

import (
	"os"

        "get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"get.porter.sh/porter/pkg/instance-storage/claimstore"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

const PluginKey = claimstore.PluginInterface + ".azure.blob"

// A sad hack because crud.Store has a method called Store which prevents us from embedding it as a field
type CrudStore = crud.Store

var _ crud.Store = &Plugin{}

// Plugin is the plugin wrapper for storing claims in azure blob storage.
type Plugin struct {
	logger hclog.Logger
	CrudStore
}

func NewPlugin(cfg azureconfig.Config) plugin.Plugin {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   PluginKey,
		Output: os.Stderr,
		Level:  hclog.Error,
	})

	return &claimstore.Plugin{
		Impl: &Plugin{
			CrudStore: NewStore(cfg, logger),
		},
	}
}
