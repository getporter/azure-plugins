package blob

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
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
		cred, err := GetCredentialsFromCli(cfg, l)
		if err != nil {
			return CredentialSet{}, errors.Errorf("environment variable %s containing the azure storage connection string was not set: %v\n%#v", credsEnv, err, cfg)
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

func GetCredentialsFromCli(cfg azureconfig.Config, l hclog.Logger) (CredentialSet, error) {

	if len(cfg.StorageAccount) == 0 || len(cfg.StorageAccountResourceGroup) == 0 {
		return CredentialSet{}, errors.Errorf("StorageAccount and/or StorageAccountResourceGroup was not set, login with az cli not attempted")
	}

	authorizer, err := auth.NewAuthorizerFromCLI()
	if err != nil {
		return CredentialSet{}, errors.Errorf("Failed to login with Azure cli: %v", err)
	}
	subscriptionId := cfg.StorageAccountSubscriptionId
	if len(subscriptionId) == 0 {
		subscriptionId, err = getCurrentAzureSubscriptionFromCli()
		if err != nil {
			return CredentialSet{}, err
		}
	}
	accountsClient := storage.NewAccountsClient(subscriptionId)
	accountsClient.Authorizer = authorizer
	_ = accountsClient.AddToUserAgent(UserAgent)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, err := accountsClient.ListKeys(ctx, cfg.StorageAccountResourceGroup, cfg.StorageAccount, "")
	if err != nil {
		return CredentialSet{}, errors.Errorf("Failed to get storage account keys: %v", err)
	}
	cred, err := azblob.NewSharedKeyCredential(cfg.StorageAccount, *(((*result.Keys)[0]).Value))
	if err != nil {
		return CredentialSet{}, errors.Errorf("Failed to create storage account credential: %v", err)
	}
	pipe := azblob.NewPipeline(cred, azblob.PipelineOptions{})
	return CredentialSet{Credential: *cred, Pipeline: pipe}, nil

}

func getCurrentAzureSubscriptionFromCli() (string, error) {
	var subscription AvailableSubscription
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Error getting home directory: %w", err)
	}
	file, err := os.Open(path.Join(home, AzureDirectory, AzureProile))
	if err != nil {
		return "", fmt.Errorf("Error getting azure profile: %w", err)
	}
	defer file.Close()

	// azureProfile can have BOM and embedded BOM so use decoder and check for BOM rather than unmarshall

	reader := bufio.NewReader(file)
	if err := removeBOM(reader); err != nil {
		return "", err
	}

	decoder := json.NewDecoder(reader)
	if _, err := decoder.Token(); err != nil {
		return "", fmt.Errorf("Error decoding opening json token: %w", err)
	}

	property, err := decoder.Token()
	if err != nil {
		return "", fmt.Errorf("Error decoding subscriptions token: %w", err)
	}
	if val, ok := property.(string); !ok || !strings.EqualFold(val, "subscriptions") {
		return "", fmt.Errorf("Error gettting subscriptions property: %w", err)
	}

	delim, err := decoder.Token()
	if err != nil {
		return "", fmt.Errorf("Error decoding json array delimiter: %w", err)
	}
	if _, ok := delim.(json.Delim); !ok {
		return "", fmt.Errorf("Error getting array delimiter: %w", err)
	}

	for decoder.More() {

		// azureProfile can have embedded BOM

		if err := removeBOM(reader); err != nil {
			return "", err
		}

		if err := decoder.Decode(&subscription); err != nil {
			return "", fmt.Errorf("Error decoding json: %w", err)
		}

		if subscription.EnvironmentName == PublicCloud && subscription.IsDefault {
			return subscription.SubscriptionId, nil
		}
	}

	return "", errors.New("Failed to get current subscription from cli config")
}

func removeBOM(reader *bufio.Reader) error {
	rune, _, err := reader.ReadRune()
	if err != nil {
		return fmt.Errorf("Error testing azure profile for BOM: %w", err)
	}
	if rune != BOM {
		if err := reader.UnreadRune(); err != nil {
			return fmt.Errorf("Failed to unread rune: %w", err)
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
