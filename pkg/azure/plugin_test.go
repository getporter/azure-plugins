package azure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	p := New()
	assert.True(t, p.IsInternalPlugin, "expected tracing to be configured as a plugin")
	assert.Equal(t, "porter.plugins.azure", p.InternalPluginKey, "expected the plugin to have its on tracing service name")
}
