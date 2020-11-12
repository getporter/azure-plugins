package azure

import (
	"strings"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"get.porter.sh/plugin/azure/pkg/azure/blob"
	"get.porter.sh/plugin/azure/pkg/azure/keyvault"
	"get.porter.sh/plugin/azure/pkg/azure/table"
	"get.porter.sh/porter/pkg/plugins"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
)

type RunOptions struct {
	Key               string
	selectedPlugin    plugin.Plugin
	selectedInterface string
}

func (o *RunOptions) Validate(args []string, cfg azureconfig.Config) error {
	if len(args) == 0 {
		return errors.New("The positional argument KEY was not specified")
	}
	if len(args) > 1 {
		return errors.New("Multiple positional arguments were specified but only one, KEY is expected")
	}

	o.Key = args[0]

	availableImplementations := getPlugins(cfg)
	selectedPlugin, ok := availableImplementations[o.Key]
	if !ok {
		return errors.Errorf("invalid plugin key specified: %q", o.Key)
	}
	o.selectedPlugin = selectedPlugin()

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
	err = opts.Validate(args, p.Config)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	plugins.Serve(opts.selectedInterface, opts.selectedPlugin)
}

func getPlugins(cfg azureconfig.Config) map[string]func() plugin.Plugin {
	return map[string]func() plugin.Plugin{
		blob.PluginInterface:     func() plugin.Plugin { return blob.NewPlugin(cfg) },
		table.PluginInterface:    func() plugin.Plugin { return table.NewPlugin(cfg) },
		keyvault.PluginInterface: func() plugin.Plugin { return keyvault.NewPlugin(cfg) },
	}
}
