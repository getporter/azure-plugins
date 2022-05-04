package azure

import (
	"strings"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"get.porter.sh/plugin/azure/pkg/azure/keyvault"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/portercontext"
	secretsplugins "get.porter.sh/porter/pkg/secrets/plugins"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
)

type RunOptions struct {
	Key               string
	selectedPlugin    pluginInitializer
	selectedInterface string
}

func (o *RunOptions) Validate(args []string) error {
	if len(args) == 0 {
		return errors.New("The positional argument KEY was not specified")
	}
	if len(args) > 1 {
		return errors.New("Multiple positional arguments were specified but only one, KEY is expected")
	}

	o.Key = args[0]

	selectedPlugin, ok := availablePlugins[o.Key]
	if !ok {
		return errors.Errorf("invalid plugin key specified: %q", o.Key)
	}
	o.selectedPlugin = selectedPlugin

	parts := strings.Split(o.Key, ".")
	o.selectedInterface = parts[0]

	return nil
}

func (p *Plugin) Run(args []string) {
	// This logger only helps log errors with loading the plugin
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "azure",
		Output:     p.Err,
		Level:      hclog.Debug,
		JSONFormat: true,
	})

	err := p.LoadConfig()
	if err != nil {
		logger.Error(err.Error())
		return
	}

	// We are not following the normal CLI pattern here because
	// if we write to stdout without the hclog, it will cause the plugin framework to blow up
	var opts RunOptions
	err = opts.Validate(args)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	plugins.Serve(p.Context, opts.selectedInterface, opts.selectedPlugin(p.Context, p.Config), secretsplugins.PluginProtocolVersion)
}

// A list of available plugins
var availablePlugins map[string]pluginInitializer = getPlugins()

type pluginInitializer func(ctx *portercontext.Context, cfg azureconfig.Config) plugin.Plugin

func getPlugins() map[string]pluginInitializer {
	return map[string]pluginInitializer{
		keyvault.PluginInterface: keyvault.NewPlugin,
	}
}
