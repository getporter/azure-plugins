//go:build mage
// +build mage

package main

import (
	// mage:import
	"get.porter.sh/magefiles/releases"
)

// We are migrating to mage, but for now keep using make as the main build script interface.

// Publish the cross-compiled binaries.
func Publish(plugin string) {
	releases.PreparePluginForPublish(plugin)
	releases.PublishPlugin(plugin)
	releases.PublishPluginFeed(plugin)
}
