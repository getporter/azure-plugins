package azure

import (
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/runtime"
)

type TestPlugin struct {
	*Plugin
	TestContext *portercontext.TestContext
}

// NewTestPlugin initializes a plugin test client, with the output buffered, and an in-memory file system.
func NewTestPlugin(t *testing.T) *TestPlugin {
	testConfig := config.NewTestConfig(t)
	m := &TestPlugin{
		Plugin: &Plugin{
			RuntimeConfig: runtime.NewConfigFor(testConfig.Config),
		},
		TestContext: testConfig.TestContext,
	}

	return m
}
