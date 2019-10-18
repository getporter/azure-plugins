package azure

import (
	"testing"

	"github.com/deislabs/porter/pkg/context"
)

type TestPlugin struct {
	*Plugin
	TestContext *context.TestContext
}

// NewTestPlugin initializes a plugin test client, with the output buffered, and an in-memory file system.
func NewTestPlugin(t *testing.T) *TestPlugin {
	c := context.NewTestContext(t)
	m := &TestPlugin{
		Plugin: &Plugin{
			Context: c.Context,
		},
		TestContext: c,
	}

	return m
}
