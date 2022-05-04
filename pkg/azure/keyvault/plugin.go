package keyvault

import (
	"os"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/secrets/plugins"
	"get.porter.sh/porter/pkg/secrets/pluginstore"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

const PluginInterface = plugins.PluginInterface + ".azure.keyvault"

var _ plugins.SecretsProtocol = &Plugin{}

// Plugin is the plugin wrapper for accessing secrets from Azure Key Vault.
type Plugin struct {
	secrets.Store
}

func NewPlugin(c *portercontext.Context, cfg azureconfig.Config) plugin.Plugin {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       PluginInterface,
		Output:     os.Stderr,
		Level:      hclog.Debug,
		JSONFormat: true,
	})

	return pluginstore.NewPlugin(c, NewStore(cfg, logger))
}
