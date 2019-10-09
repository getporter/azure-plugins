package credentials

import (
	"os"
	"regexp"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/pkg/errors"
)

type CredentialSet struct {
	Credential azblob.SharedKeyCredential
	Pipeline   pipeline.Pipeline
}

func GetCredentials() (CredentialSet, error) {
	accountName := os.Getenv("AZURE_STORAGE_ACCOUNT")
	accountKey := os.Getenv("AZURE_STORAGE_ACCESS_KEY")

	if accountName == "" || accountKey == "" {
		connString := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
		if connString == "" {
			errors.New("AZURE_STORAGE_ACCOUNT and AZURE_STORAGE_ACCESS_KEY or AZURE_STORAGE_CONNECTION_STRING must be set")
		}

		var err error
		accountName, accountKey, err = parseConnectionString(connString)
		if err != nil {
			return CredentialSet{}, err
		}
	}

	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return CredentialSet{}, err
	}
	pipe := azblob.NewPipeline(cred, azblob.PipelineOptions{})

	return CredentialSet{Credential: *cred, Pipeline: pipe}, nil
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
