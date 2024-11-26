package keyvault

import (
	"os"
	"strings"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/hashicorp/go-hclog"
)

// GetCredentials gets an authorizer for Azure
func GetCredentials(cfg azureconfig.Config, l hclog.Logger) (*azidentity.DefaultAzureCredential, error) {

	azureAuthEnvVarNames := []string{
		"AZURE_TENANT_ID",
		"AZURE_CLIENT_ID",
		"AZURE_CLIENT_SECRET",
		"AZURE_CERTIFICATE_PATH",
		"AZURE_CERTIFICATE_PASSWORD",
		"AZURE_USERNAME",
		"AZURE_PASSWORD",
	}

	prefix := cfg.EnvAzurePrefix
	if prefix != "" && prefix != "AZURE_" {
		for _, v := range azureAuthEnvVarNames {
			env := prefix + strings.TrimPrefix(v, "AZURE_")
			val := os.Getenv(env)
			os.Setenv(v, val)
		}
	}

	creds, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	return creds, nil
}
