package blob

import (
	"os"
	"strings"
	"testing"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"github.com/Azure/go-autorest/autorest/azure/cli"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetCredentials(t *testing.T) {
	testcases := []struct {
		name         string
		envVarsToSet map[string]string
		config       *azureconfig.Config
		wantError    string
	}{
		{
			"Missing Environment Variables",
			map[string]string{},
			&azureconfig.Config{},
			"environment variable AZURE_STORAGE_CONNECTION_STRING containing the azure storage connection string was not set:\nazureconfig.Config{EnvConnectionString:\"\", StorageAccount:\"\", StorageAccountResourceGroup:\"\", StorageAccountSubscriptionId:\"\", EnvAzurePrefix:\"\", Vault:\"\"}",
		},
		{
			"Invalid Connection string",
			map[string]string{
				"AZURE_STORAGE_CONNECTION_STRING": "Invalid",
			},
			&azureconfig.Config{},
			"unexpected format for AZURE_STORAGE_CONNECTION_STRING, could not find AccountName=NAME and AccountKey=KEY in it",
		},
		{
			"Valid Connection string",
			map[string]string{
				"AZURE_STORAGE_CONNECTION_STRING": "DefaultEndpointsProtocol=https;AccountName=bmFtZQo=;AccountKey=a2V5Cg==;EndpointSuffix=core.windows.net",
			},
			&azureconfig.Config{},
			"",
		},
		{
			"Missing Storage Acccount Resource Group",
			map[string]string{},
			&azureconfig.Config{
				StorageAccount: "account",
			},
			"resource-group is not set - cannot login with Azure CLI\nazureconfig.Config{EnvConnectionString:\"\", StorageAccount:\"account\", StorageAccountResourceGroup:\"\", StorageAccountSubscriptionId:\"\", EnvAzurePrefix:\"\", Vault:\"\"}",
		},
		{
			"Missing Storage Acccount Name",
			map[string]string{},
			&azureconfig.Config{
				StorageAccountResourceGroup: "group",
			},
			"account is not set - cannot login with Azure CLI\nazureconfig.Config{EnvConnectionString:\"\", StorageAccount:\"\", StorageAccountResourceGroup:\"group\", StorageAccountSubscriptionId:\"\", EnvAzurePrefix:\"\", Vault:\"\"}",
		},
	}
	for _, tc := range testcases {

		t.Run(tc.name, func(t *testing.T) {

			logger := hclog.New(&hclog.LoggerOptions{
				Name:   strings.ReplaceAll(tc.name, " ", "_"),
				Output: os.Stderr,
				Level:  hclog.Error,
			})

			for k, v := range tc.envVarsToSet {
				os.Setenv(k, v)
			}

			defer func() {
				for k := range tc.envVarsToSet {
					os.Unsetenv(k)
				}
			}()

			cred, err := GetCredentials(*tc.config, logger)
			if tc.wantError == "" {
				require.NoError(t, err, "GetCredentials should have not returned an error")
				assert.NotNil(t, cred)
			} else {
				require.Error(t, err, "GetCredentials should have returned an error")
				assert.EqualError(t, err, tc.wantError)
			}
		})
	}
}

func Test_LoginwithCLI(t *testing.T) {

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   t.Name(),
		Output: os.Stderr,
		Level:  hclog.Error,
	})

	config := &azureconfig.Config{
		StorageAccount:              "account",
		StorageAccountResourceGroup: "group",
	}

	_, err := GetCredentials(*config, logger)
	require.Error(t, err, "GetCredentials should have returned an error")
	if isLoggedInWithAzureCLI() {
		assert.Contains(t, err.Error(), "Failed to get storage account keys:")
	} else {
		assert.Contains(t, err.Error(), "Failed to login with Azure CLI:")
	}
}
func isLoggedInWithAzureCLI() bool {
	_, err := cli.GetTokenFromCLI("https://management.azure.com/")
	return err == nil
}
