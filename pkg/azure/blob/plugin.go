package blob

import (
	"os"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"get.porter.sh/porter/pkg/storage/crudstore"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

const PluginInterface = crudstore.PluginInterface + ".azure.blob"

var _ crud.Store = &Plugin{}

// Plugin is the plugin wrapper for storing claims in azure blob storage.
type Plugin struct {
	logger hclog.Logger
	crud.Store
}

func NewPlugin(cfg azureconfig.Config) plugin.Plugin {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   PluginInterface,
		Output: os.Stderr,
		Level:  hclog.Error,
	})

	return &crudstore.Plugin{
		Impl: &Plugin{
			Store: NewStore(cfg, logger),
		},
	}
}
