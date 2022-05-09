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
	"go.opentelemetry.io/otel/attribute"
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
	log.SetAttributes(attribute.String("secret name", keyValue))

	if strings.ToLower(keyName) != SecretKeyName {
		return s.hostStore.Resolve(ctx, keyName, keyValue)
	}

	if err := s.Connect(ctx); err != nil {
		return "", err
	}

	secretVersion := ""
	result, err := s.client.GetSecret(ctx, s.vaultUrl, keyValue, secretVersion)
	if err != nil {
		return "", log.Error(fmt.Errorf("could not get secret %s: %w", keyValue, err))
	}

	return *result.Value, nil
}

// Create saves a the secret to azure's keyvault using the keyValue as the
// secret key.
// It implements the Create method on the secret plugins' interface.
func (s *Store) Create(ctx context.Context, keyName string, keyValue string, value string) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	// check if the keyName is secret
	if keyName != SecretKeyName {
		return log.Error(fmt.Errorf("unsupported secret type: %s. Only %s is supported.", keyName, SecretKeyName))
	}

	if err := s.Connect(ctx); err != nil {
		return err
	}

	_, err := s.client.SetSecret(ctx, s.vaultUrl, keyValue, keyvault.SecretSetParameters{Value: &value})
	if err != nil {
		return log.Error(fmt.Errorf("failed to set secret for key %s in azure-keyvault: %w", keyValue, err))
	}

	return nil
}
