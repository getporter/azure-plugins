package common

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/pkg/errors"
)

type CredentialSet struct {
	Credential azblob.SharedKeyCredential
	Pipeline   pipeline.Pipeline
}

const PublicCloud = "AzureCloud"
const AzureDirectory = ".azure"
const AzureProile = "azureProfile.json"
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

func GetCurrentAzureSubscriptionFromCli() (string, error) {

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
	if err := RemoveBOM(reader); err != nil {
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

func RemoveBOM(reader *bufio.Reader) error {
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

func ParseConnectionString(connString string) (name string, key string, err error) {
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
