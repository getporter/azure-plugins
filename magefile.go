//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"get.porter.sh/magefiles/ci"
	"get.porter.sh/magefiles/git"
	"get.porter.sh/magefiles/porter"
	"get.porter.sh/magefiles/releases"
	"get.porter.sh/magefiles/tools"
	"github.com/carolynvs/magex/mgx"
	"github.com/carolynvs/magex/shx"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/target"
)

const (
	// Name of the plugin
	pluginName = "azure"

	// go package to build when building the plugin
	pluginPkg = "get.porter.sh/plugin/" + pluginName

	// directory where the compiled binaries for the plugin are generated
	binDir = "bin/plugins/" + pluginName
)

var (
	// Build a command that stops the build on if the command fails
	must = shx.CommandBuilder{StopOnError: true}

	// List of directories that should trigger a build when changed
	srcDirs = []string{"cmd/", "pkg/", "go.mod", "magefile.go"}
)

// Publish uploads the cross-compiled binaries for the plugin
func Publish() {
	mg.SerialDeps(porter.UseBinForPorterHome, porter.EnsurePorter)

	releases.PreparePluginForPublish(pluginName)
	releases.PublishPlugin(pluginName)
	releases.PublishPluginFeed(pluginName)
}

// TestPublish tries out publish locally, with your github forks
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

// Install the azure plugin to PORTER_HOME
func Install() {
	pluginDir := filepath.Join(porter.GetPorterHome(), "plugins", pluginName)
	mgx.Must(os.MkdirAll(pluginDir, 0700))

	// Copy the plugin into PORTER_HOME
	mgx.Must(shx.Copy(filepath.Join(binDir, pluginName), pluginDir))
}

// ConfigureAgent sets up the CI server before running the build
func ConfigureAgent() error {
	return ci.ConfigureAgent()
}

// EnsureMage installs mage
func EnsureMage() error {
	return tools.EnsureMage()
}

func Fmt() {
	must.RunV("go", "fmt", "./...")
}

func Vet() {
	must.RunV("go", "vet", "./...")
}

// Test runs all tests
func Test() {
	// Do not run TestIntegration until its safe to run in CI
	mg.SerialDeps(TestUnit)
}

// TestUnit runs the unit tests
func TestUnit() {
	v := ""
	if mg.Verbose() {
		v = "-v"
	}

	must.Command("go", "test", v, "./...").CollapseArgs().RunV()
}

// TestIntegration runs integration tests, requires AZURE_* environment variables set
// This is not yet run in CI so make sure to run locally
func TestIntegration() {
	must.RunV("bash", "./tests/integration/script.sh")
}

func Build() {
	rebuild, err := target.Dir(filepath.Join(binDir, pluginName), srcDirs...)
	if err != nil {
		mgx.Must(fmt.Errorf("error inspecting source dirs %s: %w", srcDirs, err))
	}
	if rebuild {
		mgx.Must(releases.BuildClient(pluginPkg, pluginName, binDir))
	} else {
		fmt.Println("target is up-to-date")
	}
}

// XBuildAll cross-compiles the plugin
func XBuildAll() {
	rebuild, err := target.Dir(filepath.Join(binDir, "dev/azure-linux-amd64"), srcDirs...)
	if err != nil {
		mgx.Must(fmt.Errorf("error inspecting source dirs %s: %w", srcDirs, err))
	}
	if rebuild {
		releases.XBuildAll(pluginPkg, pluginName, binDir)
	} else {
		fmt.Println("target is up-to-date")
	}

	releases.PreparePluginForPublish(pluginName)
}

func Clean() error {
	return os.RemoveAll("bin")
}

func SetupDCO() error {
	return git.SetupDCO()
}
