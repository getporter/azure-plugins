package keyvault

import (
	"io/ioutil"
	"os"
	"testing"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestResolve_NonSecret(t *testing.T) {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   PluginInterface,
		Output: os.Stderr,
		Level:  hclog.Error,
	})

	azConfig := azureconfig.Config{}
	store := NewStore(azConfig, logger)

	t.Run("resolve non-secret source: value", func(t *testing.T) {
		resolved, err := store.Resolve(host.SourceValue, "myvalue")
		require.NoError(t, err)
		require.Equal(t, "myvalue", resolved)
	})

	t.Run("resolve non-secret source: env", func(t *testing.T) {
		os.Setenv("MY_ENV_VAR", "myvalue")
		defer os.Unsetenv("MY_ENV_VAR")

		resolved, err := store.Resolve(host.SourceEnv, "MY_ENV_VAR")
		require.NoError(t, err)
		require.Equal(t, "myvalue", resolved)
	})

	t.Run("resolve non-secret source: path", func(t *testing.T) {
		file, err := ioutil.TempFile("", "myfile")
		if err != nil {
			require.NoError(t, err)
		}
		defer os.Remove(file.Name())

		_, err = file.WriteString("myfilecontents")
		require.NoError(t, err)

		resolved, err := store.Resolve(host.SourcePath, file.Name())
		require.NoError(t, err)
		require.Equal(t, "myfilecontents", resolved)
	})

	t.Run("resolve non-secret source: command", func(t *testing.T) {
		resolved, err := store.Resolve(host.SourceCommand, "echo -n Hello World!")
		require.NoError(t, err)
		require.Equal(t, "Hello World!", resolved)
	})

	t.Run("resolve non-secret source: bogus", func(t *testing.T) {
		_, err := store.Resolve("bogus", "bogus")
		require.EqualError(t, err, "invalid credential source: bogus")
	})
}
