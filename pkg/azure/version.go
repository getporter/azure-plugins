package azure

import (
	"get.porter.sh/plugin/azure/pkg"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/porter/version"
)

func (p *Plugin) PrintVersion(opts version.Options) error {
	metadata := plugins.Metadata{
		Metadata: pkgmgmt.Metadata{
			Name: "azure",
			VersionInfo: pkgmgmt.VersionInfo{
				Version: pkg.Version,
				Commit:  pkg.Commit,
				Author:  "Porter Authors",
			},
		},
		Implementations: []plugins.Implementation{
			{Type: "storage", Name: "blob"},
			{Type: "secrets", Name: "keyvault"},
		},
	}
	return version.PrintVersion(p.Context, opts, metadata)
}
