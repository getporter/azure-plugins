package blob

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strings"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"

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
const PublicCloud = "AzureCloud"
const AzureDirectory = ".azure"
const AzureProile = "azureProfile.json"
const UserAgent = "porter.azure.storage.plugin"
const BOM = '\uFEFF'

type AvailableSubscription struct {
	SubscriptionId  string `json:"id"`
	State           string `json:"state"`
	IsDefault       bool   `json:"isDefault"`
	EnvironmentName string `json:"environmentName"`
}

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
			return CredentialSet{}, errors.Errorf("%v:%#v", err, cfg)
		}
		return cred, nil
	}

	accountName, accountKey, err := parseConnectionString(connString)
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
		subscriptionId, err = getCurrentAzureSubscriptionFromCli()
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

func getCurrentAzureSubscriptionFromCli() (string, error) {
	var subscription AvailableSubscription
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "Error getting home directory")
	}
	file, err := os.Open(path.Join(home, AzureDirectory, AzureProile))
	if err != nil {
		return "", errors.Wrap(err, "Error getting azure profile")
	}
	defer file.Close()

	// azureProfile can have BOM and embedded BOM so use decoder and check for BOM rather than unmarshall

	reader := bufio.NewReader(file)
	if err := removeBOM(reader); err != nil {
		return "", err
	}

	decoder := json.NewDecoder(reader)
	if _, err := decoder.Token(); err != nil {
		return "", errors.Wrap(err, "Error decoding opening json token")
	}

	property, err := decoder.Token()
	if err != nil {
		return "", errors.Wrap(err, "Error decoding subscriptions token")
	}
	if val, ok := property.(string); !ok || !strings.EqualFold(val, "subscriptions") {
		return "", errors.Wrap(err, "Error gettting subscriptions property")
	}

	delim, err := decoder.Token()
	if err != nil {
		return "", errors.Wrap(err, "Error decoding json array delimiter")
	}
	if _, ok := delim.(json.Delim); !ok {
		return "", errors.Wrap(err, "Error getting array delimiter")
	}

	for decoder.More() {

		// azureProfile can have embedded BOM

		if err := removeBOM(reader); err != nil {
			return "", err
		}

		if err := decoder.Decode(&subscription); err != nil {
			return "", errors.Wrap(err, "Error decoding json")
		}

		if subscription.EnvironmentName == PublicCloud && subscription.IsDefault {
			return subscription.SubscriptionId, nil
		}
	}

	return "", errors.New("Failed to get current subscription from cli config")
}

func removeBOM(reader *bufio.Reader) error {
	rune, _, err := reader.ReadRune()
	if err != nil && err != io.EOF {
		return errors.Wrap(err, "Error testing azure profile for BOM")
	}
	if rune != BOM && err != io.EOF {
		if err := reader.UnreadRune(); err != nil {
			return errors.Wrap(err, "Failed to unread rune")
		}
	}
	return nil
}
func parseConnectionString(connString string) (name string, key string, err error) {
	keyRegex := regexp.MustCompile("AccountKey=([^;]+)")
	keyMatch := keyRegex.FindAllStringSubmatch(connString, -1)

	nameRegex := regexp.MustCompile("AccountName=([^;]+)")
	nameMatch := nameRegex.FindAllStringSubmatch(connString, -1)

	if len(nameMatch) == 0 || len(keyMatch) == 0 {
		return "", "", errors.New("unexpected format for AZURE_STORAGE_CONNECTION_STRING, could not find AccountName=NAME and AccountKey=KEY in it")
	}

	accountKey := keyMatch[0][1]
	accountName := nameMatch[0][1]
	return accountName, accountKey, nil
}
