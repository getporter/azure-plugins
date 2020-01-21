package azure

import (
	"get.porter.sh/plugin/azure/pkg"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/porter/version"
)

func (p *Plugin) PrintVersion(opts version.Options) error {
	metadata := mixin.Metadata{
		Name: "azure",
		VersionInfo: mixin.VersionInfo{
			Version: pkg.Version,
			Commit:  pkg.Commit,
			Author:  "Porter Authors",
		},
	}
	return version.PrintVersion(p.Context, opts, metadata)
}
