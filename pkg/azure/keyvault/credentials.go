package keyvault

import (
	"os"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/auth"
	"github.com/Azure/go-autorest/autorest"
	azureauth "github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
)

func GetCredentials(cfg azureconfig.Config, l hclog.Logger) (autorest.Authorizer, error) {
	prefix := cfg.EnvServicePrincipalPrefix
	if prefix != "" && prefix != "AZURE_" {
		tenantEnv := prefix + "TENANT_ID"
		clientEnv := prefix + "CLIENT_ID"
		secretEnv := prefix + "CLIENT_SECRET"

		tenant := os.Getenv(tenantEnv)
		client := os.Getenv(clientEnv)
		secret := os.Getenv(secretEnv)

		// NewAuthorizerFromEnvironment only reads from well-known env vars
		os.Setenv(azureauth.TenantID, tenant)
		os.Setenv(azureauth.ClientID, client)
		os.Setenv(azureauth.ClientSecret, secret)
	}

	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create an azure authorizer")
	}

	return authorizer, nil
}
