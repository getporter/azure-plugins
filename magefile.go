//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"

	// mage:import
	"get.porter.sh/magefiles/releases"
)

const (
	// Name of the plugin
	pluginName = "azure"
)

// Publish the cross-compiled binaries.
func Publish() {
	releases.PreparePluginForPublish(pluginName)
	releases.PublishPlugin(pluginName)
	releases.PublishPluginFeed(pluginName)
}

// Test out publish locally, with your github forks
// Assumes that you forked and kept the repository name unchanged.
func TestPublish(username string) {
	pluginRepo := fmt.Sprintf("github.com/%s/%s-plugins", username, pluginName)
	pkgRepo := fmt.Sprintf("https://github.com/%s/packages.git", username)
	fmt.Printf("Publishing a release to %s and committing a mixin feed to %s\n", pluginRepo, pkgRepo)
	fmt.Printf("If you use different repository names, set %s and %s then call mage Publish instead.\n", releases.ReleaseRepository, releases.PackagesRemote)
	os.Setenv(releases.ReleaseRepository, pluginRepo)
	os.Setenv(releases.PackagesRemote, pkgRepo)

	Publish()
}
