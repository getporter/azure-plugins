package blob

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"

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

type AvailableSubscriptions struct {
	Subscriptions []AvailableSubscription `json:"subscriptions"`
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
			return CredentialSet{}, errors.Errorf("%v\n%#v", err, cfg)
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

	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "Error getting home directory")
	}

	return getCurrentAzureSubscriptionFromProfile(path.Join(home, AzureDirectory, AzureProile))
}

func getCurrentAzureSubscriptionFromProfile(filename string) (string, error) {

	file, err := os.Open(filename)
	if err != nil {
		return "", errors.Wrap(err, "Error getting azure profile")
	}
	defer file.Close()

	// azureProfile can have BOM so check for BOM before decoding

	reader := bufio.NewReader(file)
	if err := removeBOM(reader); err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", errors.Wrap(err, "Error reading Azure profile")
	}

	var subscriptions AvailableSubscriptions
	if err := json.Unmarshal(data, &subscriptions); err != nil {
		return "", errors.Wrap(err, "Failed to decode Azure Profile")
	}

	for _, availableSubscription := range subscriptions.Subscriptions {
		if availableSubscription.EnvironmentName == PublicCloud && availableSubscription.IsDefault {
			return availableSubscription.SubscriptionId, nil
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
