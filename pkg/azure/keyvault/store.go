package keyvault

import (
	"context"
	"crypto/md5"
	"fmt"
	"net/url"
	"regexp"
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

type secret struct {
	vaultURL string
	name     string
	version  string
}

// Store implements the backing store for secrets in azure key vault.
type Store struct {
	logger    hclog.Logger
	config    azureconfig.Config
	vaultUrl  string
	client    *keyvault.BaseClient
	hostStore host.Store
}

func NewStore(cfg azureconfig.Config, l hclog.Logger) *Store {
	vaultFullLink := cfg.VaultUrl
	if vaultFullLink == "" {
		vaultFullLink = fmt.Sprintf("https://%s.vault.azure.net", cfg.Vault)
	}
	
	return &Store{
		config:    cfg,
		logger:    l,
		vaultUrl:  vaultFullLink,
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

	log.SetAttributes(attribute.String("requested-secret", keyValue))

	if err := s.Connect(ctx); err != nil {
		return "", err
	}
	// Check if the keyValue is set to a full ID or just the secret name. The keyValue is only considered
	// an ID if it includes at least the keyvault name and secret name. If version is not part of the ID then the version
	// is set to "" which will fetch the latest version
	secret := parseID(ctx, keyValue)
	if secret != nil {
		result, err := s.client.GetSecret(ctx, secret.vaultURL, secret.name, secret.version)
		if err != nil {
			// Instead of return error in this case instead log as a debug and attempt to fetch
			// the secret from the configured secret store. Only return error if the secret is unable
			// to be resolved in both ways
			log.Debug(fmt.Sprintf("could not get secret %s by ID: %s", keyValue, err.Error()))
		} else {
			// If we were able to look it up based off of the parsed ID then return that immediately
			return *result.Value, nil
		}
	}

	secretName := cleanSecretName(keyValue)
	attribute.String("cleaned-secret", secretName)

	secretVersion := ""
	result, err := s.client.GetSecret(ctx, s.vaultUrl, secretName, secretVersion)
	if err != nil {
		if keyValue != secretName {
			// Help everyone out by printing the original value that we used to generate the secret name
			return "", log.Errorf("could not get secret %s (original name was %s): %w", secretName, keyValue, err)
		}
		return "", log.Errorf("could not get secret %s: %w", secretName, err)
	}

	return *result.Value, nil
}

// Matches any invalid characters in an Azure Key Vault name so that we can replace it with something allowed
var keyVaultNameInvalidCharacters = regexp.MustCompile(`[^a-zA-Z0-9-]`)

// cleanSecretName replaces any invalid characters in the secret name with a
// hyphen and ensures that the name is 127 characters or fewer. When it's too
// long, we generate a md5 sum of the original name and append it to as much of
// the cleaned up name as we can preserve.
//
// We need this because Porter supports a larger set of parameter name characters
// than Azure Key Vault, which only allows alphanumeric characters and hyphens.
// Example: MY_SECRET is converted to MY-SECRET when read/written to key vault
// or INSTALLATION-ID-LONG-SECRET-NAME is converted to INSTALLATION-ID-CLEAN_SECRET_PREFIX-MD5SUM
func cleanSecretName(name string) string {
	cleanName := keyVaultNameInvalidCharacters.ReplaceAllString(name, "-")
	if len(cleanName) > 127 {
		// If the name is too long, hash the original and append the hash to as much of the name as we can preserve
		nameHash := fmt.Sprintf("%X", md5.Sum([]byte(name)))
		cleanName = cleanName[:94] + "-" + nameHash
	}

	return cleanName
}

// Create saves the secret to azure's keyvault using the keyValue as the
// secret key.
// It implements the Create method on the secret plugins' interface.
func (s *Store) Create(ctx context.Context, keyName string, keyValue string, value string) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	// check if the keyName is secret
	if keyName != SecretKeyName {
		return log.Errorf("unsupported secret type: %s. Only %s is supported", keyName, SecretKeyName)
	}

	secretName := cleanSecretName(keyValue)
	log.SetAttributes(
		attribute.String("requested-secret", keyValue),
		attribute.String("cleaned-secret", secretName))

	if err := s.Connect(ctx); err != nil {
		return err
	}

	_, err := s.client.SetSecret(ctx, s.vaultUrl, secretName, keyvault.SecretSetParameters{Value: &value})
	if err != nil {
		if keyValue != secretName {
			// Help everyone out by printing the original value that we used to generate the secret name
			return log.Errorf("failed to set secret %s (original name was %s): %w", secretName, keyValue, err)
		}
		return log.Errorf("failed to set secret %s in azure-keyvault: %w", secretName, err)
	}
	return nil
}

// parseID will attempt to create a secret from an id. If the id is not valid then
// it will log a debug and return nil. This code was mainly copied from the azure keyvault internal library:
// https://github.com/Azure/azure-sdk-for-go/blob/main/sdk/keyvault/internal/parse.go
func parseID(ctx context.Context, id string) *secret {
	_, log := tracing.StartSpan(ctx, attribute.String("parsing secret as ID", id))
	if id == "" {
		log.Debug("unable to parse empty ID")
		return nil
	}
	parsed, err := url.Parse(id)
	if err != nil {
		log.Debug(fmt.Sprintf("Unable to parse %s as secret ID: %s", id, err.Error()))
		return nil
	}
	url := fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
	// Trim preceeding and trailing slashes
	split := strings.Split(strings.TrimSuffix(strings.TrimPrefix(parsed.Path, "/"), "/"), "/")
	if len(split) < 3 {
		if len(split) == 2 {
			return &secret{
				vaultURL: url,
				name:     split[1],
				version:  "",
			}
		}
		log.Debug(fmt.Sprintf("Unexpected ID format found for %s, unable to parse as secret ID", id))
		return nil
	}
	return &secret{
		vaultURL: url,
		name:     split[1],
		version:  split[2],
	}
}
