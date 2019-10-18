package azure

import (
	"strings"

	"github.com/deislabs/porter/pkg/plugins"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
)

type RunOptions struct {
	Implementation    string
	selectedPlugin    plugin.Plugin
	selectedInterface string
}

func (o *RunOptions) Validate(args []string) error {
	if len(args) == 0 {
		return errors.New("The positional argument PLUGIN_IMPLEMENTATION was not specified")
	}
	if len(args) > 1 {
		return errors.New("Multiple positional arguments were specified but only one, PLUGIN_IMPLEMENTATION is expected")
	}

	o.Implementation = args[0]

	availableImplementations := GetAvailableImplementations()
	selectedPlugin, ok := availableImplementations[o.Implementation]
	if !ok {
		return errors.Errorf("invalid PLUGIN_IMPLEMENTATION specified: %q", o.Implementation)
	}
	o.selectedPlugin = selectedPlugin()

	parts := strings.Split(o.Implementation, ".")
	o.selectedInterface = parts[0]

	return nil
}

func (p *Plugin) Run(opts RunOptions) {
	plugins.Serve(opts.selectedInterface, opts.selectedPlugin)
}
