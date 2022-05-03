package keyvault

import (
	"context"
	"fmt"
	"strings"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"get.porter.sh/porter/pkg/secrets/plugins"
	"get.porter.sh/porter/pkg/secrets/plugins/host"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/v7.0/keyvault"
	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
)

var _ plugins.SecretsProtocol = &Store{}

const (
	SecretKeyName = "secret"
)

// Store implements the backing store for secrets in azure key vault.
type Store struct {
	logger    hclog.Logger
	config    azureconfig.Config
	vaultUrl  string
	client    *keyvault.BaseClient
	hostStore host.Store
}

func NewStore(cfg azureconfig.Config, l hclog.Logger) *Store {
	return &Store{
		config:    cfg,
		logger:    l,
		vaultUrl:  fmt.Sprintf("https://%s.vault.azure.net", cfg.Vault),
		hostStore: host.NewStore(),
	}
}

func (s *Store) Connect(ctx context.Context) error {
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

func (s *Store) Resolve(ctx context.Context, keyName string, keyValue string) (string, error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	if strings.ToLower(keyName) != SecretKeyName {
		return s.hostStore.Resolve(ctx, keyName, keyValue)
	}

	if err := s.Connect(ctx); err != nil {
		return "", err
	}

	secretVersion := ""
	result, err := s.client.GetSecret(context.Background(), s.vaultUrl, keyValue, secretVersion)
	if err != nil {
		return "", errors.Wrapf(err, "could not get secret %s from %s", keyValue, s.vaultUrl)
	}

	return *result.Value, nil
}

// Create implements the Create method on the secret plugins' interface.
func (s *Store) Create(ctx context.Context, keyName string, keyValue string, value string) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return err
	}

	// check if the keyName is secret
	if keyName != SecretKeyName {
		return log.Error(errors.New("invalid key name: " + keyName))
	}

	_, err := s.client.SetSecret(ctx, s.vaultUrl, keyValue, keyvault.SecretSetParameters{Value: &value})
	if err != nil {
		return log.Error(fmt.Errorf("failed to create key: %s: %w", keyName, err))
	}

	return nil

}
