package keyvault

import (
	"fmt"
	"os"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/secrets/plugins"
	"get.porter.sh/porter/pkg/secrets/pluginstore"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/mitchellh/mapstructure"
)

const PluginInterface = plugins.PluginInterface + ".azure.keyvault"

var _ plugins.SecretsProtocol = &Plugin{}

// Plugin is the plugin wrapper for accessing secrets from Azure Key Vault.
type Plugin struct {
	secrets.Store
}

func NewPlugin(c *portercontext.Context, rawCfg interface{}) (plugin.Plugin, error) {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       PluginInterface,
		Output:     os.Stderr,
		Level:      hclog.Debug,
		JSONFormat: true,
	})

	cfg := azureconfig.Config{}
	if err := mapstructure.Decode(rawCfg, &cfg); err != nil {
		return nil, fmt.Errorf("error reading plugin configuration: %w", err)
	}
	impl := NewStore(cfg, logger)

	return pluginstore.NewPlugin(c, impl)
}
