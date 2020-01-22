package azure

import (
	"get.porter.sh/plugin/azure/pkg/azure/blob"
	"get.porter.sh/porter/pkg/context"
	"github.com/hashicorp/go-plugin"
)

type Plugin struct {
	*context.Context
}

// New azure plugin client, initialized with useful defaults.
func New() *Plugin {
	return &Plugin{
		Context: context.New(),
	}
}

func GetAvailableImplementations() map[string]func() plugin.Plugin {
	return map[string]func() plugin.Plugin{
		blob.PluginKey: blob.NewPlugin,
	}
}
