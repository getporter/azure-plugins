package keyvault

import (
	"os"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"get.porter.sh/porter/pkg/secrets"
	cnabsecrets "github.com/cnabio/cnab-go/secrets"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

const PluginInterface = secrets.PluginInterface + ".azure.keyvault"

var _ cnabsecrets.Store = &Plugin{}

// Plugin is the plugin wrapper for accessing secrets from Azure Key Vault.
type Plugin struct {
	logger hclog.Logger
	cnabsecrets.Store
}

func NewPlugin(cfg azureconfig.Config) plugin.Plugin {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   PluginInterface,
		Output: os.Stderr,
		Level:  hclog.Error,
	})

	return &secrets.Plugin{
		Impl: &Plugin{
			Store: NewStore(cfg, logger),
		},
	}
}
