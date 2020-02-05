package keyvault

import (
	"context"
	"fmt"
	"strings"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"get.porter.sh/porter/pkg/secrets"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/v7.0/keyvault"
	cnabsecrets "github.com/cnabio/cnab-go/secrets"
	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
)

var _ cnabsecrets.Store = &Store{}

const (
	SecretKeyName = "secret"
)

// Store implements the backing store for secrets in azure key vault.
type Store struct {
	logger   hclog.Logger
	config   azureconfig.Config
	vaultUrl string
	client   *keyvault.BaseClient
}

func NewStore(cfg azureconfig.Config, l hclog.Logger) cnabsecrets.Store {
	s := &Store{
		config:   cfg,
		logger:   l,
		vaultUrl: fmt.Sprintf("https://%s.vault.azure.net", cfg.Vault),
	}

	return secrets.NewSecretStore(s)
}

func (s *Store) Connect() error {
	if s.client != nil {
		return nil
	}

	authorizer, err := GetCredentials(s.config, s.logger)
	if err != nil {
		return err
	}

	client := keyvault.New()
	s.client = &client
	s.client.Authorizer = authorizer
	return nil
}

func (s *Store) Resolve(keyName string, keyValue string) (string, error) {
	if strings.ToLower(keyName) != SecretKeyName {
		return "", errors.Errorf("cannot resolve unsupported keyName: %s. The azure key vault plugin only supports '%s' right now", keyName, SecretKeyName)
	}

	secretVersion := ""
	result, err := s.client.GetSecret(context.Background(), s.vaultUrl, keyValue, secretVersion)
	if err != nil {
		return "", errors.Wrapf(err, "could not get secret %s from %s", keyValue, s.vaultUrl)
	}

	return *result.Value, nil
}
