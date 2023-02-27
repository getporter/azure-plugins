package keyvault

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve_NonSecret(t *testing.T) {

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   PluginInterface,
		Output: os.Stderr,
		Level:  hclog.Error,
	})
	ctx := context.Background()

	azConfig := azureconfig.Config{}
	store := NewStore(azConfig, logger)
	os.Setenv("AZURE_CLIENT_ID", "test_client_id")
	defer os.Unsetenv("AZURE_CLIENT_ID")
	t.Run("resolve non-secret source: value", func(t *testing.T) {
		resolved, err := store.Resolve(ctx, host.SourceValue, "myvalue")
		require.NoError(t, err)
		require.Equal(t, "myvalue", resolved)
	})

	t.Run("resolve non-secret source: env", func(t *testing.T) {
		os.Setenv("MY_ENV_VAR", "myvalue")
		defer os.Unsetenv("MY_ENV_VAR")

		resolved, err := store.Resolve(ctx, host.SourceEnv, "MY_ENV_VAR")
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

		resolved, err := store.Resolve(ctx, host.SourcePath, file.Name())
		require.NoError(t, err)
		require.Equal(t, "myfilecontents", resolved)
	})

	t.Run("resolve non-secret source: command", func(t *testing.T) {
		resolved, err := store.Resolve(ctx, host.SourceCommand, "echo -n Hello World!")
		require.NoError(t, err)
		require.Equal(t, "Hello World!", resolved)
	})

	t.Run("resolve non-secret source: bogus", func(t *testing.T) {
		_, err := store.Resolve(ctx, "bogus", "bogus")
		require.EqualError(t, err, "invalid value source: bogus")
	})
}

func TestParseKeyValueAsSecretID(t *testing.T) {
	tests := []struct {
		name     string
		keyValue string
		exp      *secret
	}{
		{
			name:     "KeyValueValidSecretID",
			keyValue: "https://myvaultname.vault.azure.net/secrets/my-secret/b86c2e6ad9054f4abf69cc185b99aa60",
			exp: &secret{
				vaultURL: "https://myvaultname.vault.azure.net",
				name:     "my-secret",
				version:  "b86c2e6ad9054f4abf69cc185b99aa60",
			},
		},
		{
			name:     "KeyValueDoesNotIncludeVersion",
			keyValue: "https://myvaultname.vault.azure.net/secrets/my-secret",
			exp: &secret{
				vaultURL: "https://myvaultname.vault.azure.net",
				name:     "my-secret",
				version:  "",
			},
		},
		{
			name:     "KeyValueHasEmptyVersion",
			keyValue: "https://myvaultname.vault.azure.net/secrets/my-secret/",
			exp: &secret{
				vaultURL: "https://myvaultname.vault.azure.net",
				name:     "my-secret",
				version:  "",
			},
		},
		{
			name:     "KeyValueMissingSecret",
			keyValue: "https://myvaultname.vault.azure.net/secrets/",
			exp:      nil,
		},
		{
			name:     "KeyValueIsInvalidURL",
			keyValue: "test:/?not-keyvault",
			exp:      nil,
		},
		{
			name:     "KeyValueIsSecretNameOnly",
			keyValue: "my-secret",
			exp:      nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			got := parseID(ctx, test.keyValue)
			require.Equal(t, test.exp, got)
		})
	}
}

func TestCleanSecretName(t *testing.T) {
	testcases := map[string]string{
		// valid characters
		"MY_Secret0": "MY-Secret0",
		// repeated invalid characters
		"My-__Secret.1": "My---Secret-1",
		// more invalid characters
		"My$Secret9": "My-Secret9",
		// spaces
		"My Secret1": "My-Secret1",
		// longer than 127 characters
		"INSTALLATION-ID-Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua": "INSTALLATION-ID-Lorem-ipsum-dolor-sit-amet--consectetur-adipiscing-elit--sed-do-eiusmod-tempor-355D661555999117D32FEC8D37E6F14E",
	}

	for input, wantOutput := range testcases {
		t.Run(input, func(t *testing.T) {
			gotOutput := cleanSecretName(input)
			assert.Equal(t, wantOutput, gotOutput, "Invalid clean name %s for %s, expected %s", gotOutput, input, wantOutput)
		})
	}
}
