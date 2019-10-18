package azure

import (
	"github.com/deislabs/porter-azure-plugins/pkg"
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/deislabs/porter/pkg/porter/version"
)

func (p *Plugin) PrintVersion(opts version.Options) error {
	metadata := mixin.Metadata{
		Name: "azure",
		VersionInfo: mixin.VersionInfo{
			Version: pkg.Version,
			Commit:  pkg.Commit,
			Author:  "DeisLabs",
		},
	}
	return version.PrintVersion(p.Context, opts, metadata)
}
