package azure

import (
	"strings"
	"testing"

	"get.porter.sh/plugin/azure/pkg"
	"get.porter.sh/porter/pkg/porter/version"
	"get.porter.sh/porter/pkg/printer"
	"github.com/stretchr/testify/require"
)

func TestPrintVersion(t *testing.T) {
	pkg.Commit = "abc123"
	pkg.Version = "v1.2.3"

	m := NewTestPlugin(t)

	opts := version.Options{}
	err := opts.Validate()
	require.NoError(t, err)
	m.PrintVersion(opts)

	gotOutput := m.TestContext.GetOutput()
	wantOutput := "azure v1.2.3 (abc123) by Porter Authors"
	if !strings.Contains(gotOutput, wantOutput) {
		t.Fatalf("invalid output:\nWANT:\t%q\nGOT:\t%q\n", wantOutput, gotOutput)
	}
}

func TestPrintJsonVersion(t *testing.T) {
	pkg.Commit = "abc123"
	pkg.Version = "v1.2.3"

	m := NewTestPlugin(t)

	opts := version.Options{}
	opts.RawFormat = string(printer.FormatJson)
	err := opts.Validate()
	require.NoError(t, err)
	m.PrintVersion(opts)

	gotOutput := m.TestContext.GetOutput()
	wantOutput := `{
  "name": "azure",
  "version": "v1.2.3",
  "commit": "abc123",
  "author": "Porter Authors",
  "implementations": [
    {
      "type": "storage",
      "implementation": "blob"
    },
    {
      "type": "secrets",
      "implementation": "keyvault"
    }
  ]
}`
	if !strings.Contains(gotOutput, wantOutput) {
		t.Fatalf("invalid output:\nWANT:\t%q\nGOT:\t%q\n", wantOutput, gotOutput)
	}
}
