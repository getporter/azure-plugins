package azure

import (
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/deislabs/porter/pkg/plugins"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
)

type RunOptions struct {
	Key               string
	selectedPlugin    plugin.Plugin
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

	availableImplementations := GetAvailableImplementations()
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
	// We are not following the normal CLI pattern here because
	// if we write to stdout without the hclog, it will cause the plugin framework to blow up
	var opts RunOptions
	err := opts.Validate(args)
	if err != nil {
		logger := hclog.New(&hclog.LoggerOptions{
			Name:   "azure",
			Output: p.Err,
			Level:  hclog.Error,
		})
		logger.Error(err.Error())
		return
	}

	plugins.Serve(opts.selectedInterface, opts.selectedPlugin)
}
