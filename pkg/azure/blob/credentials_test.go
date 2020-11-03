package blob

import (
	"os"
	"strings"
	"testing"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

func TestGet_GetCredentials(t *testing.T) {
	testcases := []struct {
		name         string
		envVarsToSet map[string]string
		config       *azureconfig.Config
		testfunc     func(t *testing.T, envVarsToSet map[string]string, config *azureconfig.Config, logger hclog.Logger)
	}{
		{
			"Missing Environment Variables",
			map[string]string{},
			&azureconfig.Config{},
			missingEnvironmentVariables,
		},
		{
			"Invalid Connection string",
			map[string]string{
				"AZURE_STORAGE_CONNECTION_STRING": "Invalid",
			},
			&azureconfig.Config{},
			invalidConnectionString,
		},
		{
			"Valid Connection string",
			map[string]string{
				"AZURE_STORAGE_CONNECTION_STRING": "DefaultEndpointsProtocol=https;AccountName=bmFtZQo=;AccountKey=a2V5Cg==;EndpointSuffix=core.windows.net",
			},
			&azureconfig.Config{},
			validConnectionString,
		},
		{
			"Missing Storage Acccount Resource Group",
			map[string]string{},
			&azureconfig.Config{
				StorageAccount: "account",
			},
			missingStorageAccountResourceGroup,
		},
		{
			"Missing Storage Acccount Name",
			map[string]string{},
			&azureconfig.Config{
				StorageAccountResourceGroup: "group",
			},
			missingStorageAccountName,
		},
		{
			"loginwithCLI",
			map[string]string{},
			&azureconfig.Config{
				StorageAccount:              "account",
				StorageAccountResourceGroup: "group",
			},
			loginwithCLI,
		},
	}
	env := os.Environ()
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

			tc.testfunc(t, tc.envVarsToSet, tc.config, logger)
			resetEnvironmentVars(t, env)
		})
	}
}

func resetEnvironmentVars(t *testing.T, env []string) {
	os.Clearenv()
	for _, e := range env {
		pair := strings.Split(e, "=")
		t.Logf("Resetting Env Variable: %s", pair[0])
		os.Setenv(pair[0], pair[1])
	}
}

func missingEnvironmentVariables(t *testing.T, envVarsToSet map[string]string, config *azureconfig.Config, logger hclog.Logger) {
	_, err := GetCredentials(*config, logger)
	assert.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "environment variable AZURE_STORAGE_CONNECTION_STRING containing the azure storage connection string was not set: StorageAccount and/or StorageAccountResourceGroup was not set, login with az cli not attempted"))
}

func invalidConnectionString(t *testing.T, envVarsToSet map[string]string, config *azureconfig.Config, logger hclog.Logger) {
	_, err := GetCredentials(*config, logger)
	assert.EqualError(t, err, "unexpected format for AZURE_STORAGE_CONNECTION_STRING, could not find AccountName=NAME and AccountKey=KEY in it")
}

func validConnectionString(t *testing.T, envVarsToSet map[string]string, config *azureconfig.Config, logger hclog.Logger) {
	_, err := GetCredentials(*config, logger)
	assert.NoError(t, err)
}

func missingStorageAccountResourceGroup(t *testing.T, envVarsToSet map[string]string, config *azureconfig.Config, logger hclog.Logger) {
	_, err := GetCredentials(*config, logger)
	assert.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "environment variable AZURE_STORAGE_CONNECTION_STRING containing the azure storage connection string was not set: StorageAccount and/or StorageAccountResourceGroup was not set, login with az cli not attempted"))
}

func missingStorageAccountName(t *testing.T, envVarsToSet map[string]string, config *azureconfig.Config, logger hclog.Logger) {
	_, err := GetCredentials(*config, logger)
	assert.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "environment variable AZURE_STORAGE_CONNECTION_STRING containing the azure storage connection string was not set: StorageAccount and/or StorageAccountResourceGroup was not set, login with az cli not attempted"))
}

func loginwithCLI(t *testing.T, envVarsToSet map[string]string, config *azureconfig.Config, logger hclog.Logger) {
	_, err := GetCredentials(*config, logger)
	if err != nil {
		assert.True(t, strings.HasPrefix(err.Error(), "environment variable AZURE_STORAGE_CONNECTION_STRING containing the azure storage connection string was not set: Failed to get storage account keys:") || strings.HasPrefix(err.Error(), "environment variable AZURE_STORAGE_CONNECTION_STRING containing the azure storage connection string was not set: Failed to login with Azure cli:"))
		return
	}
}
