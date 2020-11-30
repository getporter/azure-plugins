package table

import (
	"context"
	"fmt"
	"os"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"get.porter.sh/plugin/azure/pkg/azure/common"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/storage/mgmt/storage"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
)

const ConnectionEnvironmentVariable = "AZURE_STORAGE_CONNECTION_STRING"
const UserAgent = "porter.azure.storage.plugin"

func GetCredentials(cfg azureconfig.Config, l hclog.Logger) (string, string, error) {
	var credsEnv = cfg.EnvConnectionString
	if credsEnv == "" {
		credsEnv = ConnectionEnvironmentVariable
	}

	connString := os.Getenv(credsEnv)
	if connString == "" {
		accountName, accountKey, useCli, err := GetCredentialsFromCli(cfg, l)
		if !useCli {
			return "", "", errors.Errorf("environment variable %s containing the azure storage connection string was not set:\n%#v", credsEnv, cfg)
		}
		if err != nil {
			return "", "", errors.Errorf("%v\n%#v", err, cfg)
		}
		return accountName, accountKey, nil
	}

	accountName, accountKey, err := common.ParseConnectionString(connString)
	if err != nil {
		return "", "", err
	}

	return accountName, accountKey, nil
}

func GetCredentialsFromCli(cfg azureconfig.Config, l hclog.Logger) (string, string, bool, error) {

	if cfg.StorageAccount == "" && cfg.StorageAccountResourceGroup == "" {
		return "", "", false, nil
	}

	if cfg.StorageAccount == "" {
		return "", "", true, errors.New("account is not set - cannot login with Azure CLI")
	}

	if cfg.StorageAccountResourceGroup == "" {
		return "", "", true, errors.New("resource-group is not set - cannot login with Azure CLI")
	}

	authorizer, err := auth.NewAuthorizerFromCLI()
	if err != nil {
		return "", "", true, errors.Errorf("Failed to login with Azure CLI: %v", err)
	}
	subscriptionId := cfg.StorageAccountSubscriptionId
	if subscriptionId == "" {
		subscriptionId, err = common.GetCurrentAzureSubscriptionFromCli()
		if err != nil {
			return "", "", true, err
		}
	}
	accountsClient := storage.NewAccountsClient(subscriptionId)
	accountsClient.Authorizer = authorizer
	err = accountsClient.AddToUserAgent(UserAgent)
	if err != nil {
		l.Debug(fmt.Sprintf("Error updating User Agent string for Azure: %v", err))
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, err := accountsClient.ListKeys(ctx, cfg.StorageAccountResourceGroup, cfg.StorageAccount, "")
	if err != nil {
		return "", "", true, errors.Errorf("Failed to get storage account keys: %v", err)
	}
	storageAccountKey := (*result.Keys)[0]
	return cfg.StorageAccount, *storageAccountKey.Value, true, nil

}
