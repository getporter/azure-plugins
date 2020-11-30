package blob

import (
	"context"
	"fmt"
	"os"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"get.porter.sh/plugin/azure/pkg/azure/common"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/storage/mgmt/storage"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
)

type CredentialSet struct {
	Credential azblob.SharedKeyCredential
	Pipeline   pipeline.Pipeline
}

const ConnectionEnvironmentVariable = "AZURE_STORAGE_CONNECTION_STRING"
const UserAgent = "porter.azure.storage.plugin"

func GetCredentials(cfg azureconfig.Config, l hclog.Logger) (CredentialSet, error) {
	var credsEnv = cfg.EnvConnectionString
	if credsEnv == "" {
		credsEnv = ConnectionEnvironmentVariable
	}

	connString := os.Getenv(credsEnv)
	if connString == "" {
		cred, useCli, err := GetCredentialsFromCli(cfg, l)
		if !useCli {
			return CredentialSet{}, errors.Errorf("environment variable %s containing the azure storage connection string was not set:\n%#v", credsEnv, cfg)
		}
		if err != nil {
			return CredentialSet{}, errors.Errorf("%v\n%#v", err, cfg)
		}
		return cred, nil
	}

	accountName, accountKey, err := common.ParseConnectionString(connString)
	if err != nil {
		return CredentialSet{}, err
	}

	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return CredentialSet{}, err
	}
	pipe := azblob.NewPipeline(cred, azblob.PipelineOptions{})

	return CredentialSet{Credential: *cred, Pipeline: pipe}, nil
}

func GetCredentialsFromCli(cfg azureconfig.Config, l hclog.Logger) (CredentialSet, bool, error) {

	if cfg.StorageAccount == "" && cfg.StorageAccountResourceGroup == "" {
		return CredentialSet{}, false, nil
	}

	if cfg.StorageAccount == "" {
		return CredentialSet{}, true, errors.New("account is not set - cannot login with Azure CLI")
	}

	if cfg.StorageAccountResourceGroup == "" {
		return CredentialSet{}, true, errors.New("resource-group is not set - cannot login with Azure CLI")
	}

	authorizer, err := auth.NewAuthorizerFromCLI()
	if err != nil {
		return CredentialSet{}, true, errors.Wrap(err, "Failed to login with Azure CLI")
	}
	subscriptionId := cfg.StorageAccountSubscriptionId
	if subscriptionId == "" {
		subscriptionId, err = common.GetCurrentAzureSubscriptionFromCli()
		if err != nil {
			return CredentialSet{}, true, err
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
		return CredentialSet{}, true, errors.Wrap(err, "Failed to get storage account keys")
	}
	storageAccountKey := (*result.Keys)[0]
	cred, err := azblob.NewSharedKeyCredential(cfg.StorageAccount, *storageAccountKey.Value)
	if err != nil {
		return CredentialSet{}, true, errors.Wrap(err, "Failed to create storage account credential")
	}
	pipe := azblob.NewPipeline(cred, azblob.PipelineOptions{})
	return CredentialSet{Credential: *cred, Pipeline: pipe}, true, nil

}
