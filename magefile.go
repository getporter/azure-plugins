//go:build mage
// +build mage

package main

import (
	// mage:import
	"get.porter.sh/magefiles/releases"
)

func Publish(plugin string) {
	releases.PreparePluginForPublish(plugin)
	releases.PublishPlugin(plugin)
	releases.PublishPluginFeed(plugin)
}
