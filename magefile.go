// +build mage

package main

import (
	// mage:import
	"get.porter.sh/porter/mage/releases"
)

// We are migrating to mage, but for now keep using make as the main build script interface.

// Publish the cross-compiled binaries.
func Publish(plugin string, version string, permalink string) {
	releases.PreparePluginForPublish(plugin, version, permalink)
	releases.PublishPlugin(plugin, version, permalink)
	releases.PublishPluginFeed(plugin, version)
}
