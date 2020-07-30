package keyvault

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/auth"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	azureauth "github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
)

// GetCredentials gets an authorizer for Azure
func GetCredentials(cfg azureconfig.Config, l hclog.Logger) (autorest.Authorizer, error) {

	azureAuthEnvVarNames := []string{
		azureauth.TenantID,
		azureauth.ClientID,
		azureauth.ClientSecret,
		azureauth.CertificatePath,
		azureauth.CertificatePassword,
		azureauth.Username,
		azureauth.Password,
	}

	prefix := cfg.EnvAzurePrefix
	if prefix != "" && prefix != "AZURE_" {
		for _, v := range azureAuthEnvVarNames {
			env := prefix + strings.TrimPrefix(v, "AZURE_")
			val := os.Getenv(env)
			os.Setenv(v, val)
		}
	}

	usedeviceCode, _ := strconv.ParseBool(cfg.LoginWithDeviceCode)
	useMSI, _ := strconv.ParseBool(cfg.LoginWithMSI)
	noVarsAreSet := noAzureAuthEnvVarsAreSet(azureAuthEnvVarNames)

	if err := validateOptions(usedeviceCode, useMSI, noVarsAreSet, prefix); err != nil {
		return nil, err
	}

	var authorizer autorest.Authorizer
	var err error

	// 1. Attempt to login with az cli or MSI if no vars are set.

	if noVarsAreSet {
		if useMSI {
			// NewAuthorizierFromEnvironment attempts to authenticate using credentials, then certicates, then user name and password and then MSI
			// If no AZURE_* envvars are set then it fall through to use MSI
			authorizer, err = auth.NewAuthorizerFromEnvironment()
			if err != nil {
				return nil, errors.Wrap(err, "Failed to create an azure authorizer from environment")
			}
		} else {
			authorizer, err = auth.NewAuthorizerFromCLI()
			if err != nil {
				return nil, errors.Wrap(err, "Failed to create an azure authorizer from azure cli")
			}
		}

		return authorizer, nil
	}

	resourceID, err := getAzureKeyVaultResourceID()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get Azure Key Vault Resource ID")
	}

	var tenantID = os.Getenv(azureauth.TenantID)

	// 2. Attempt to login with Device Code - device code requires an appid, we should create a fixed appId for the plugin but for now we can get this from an env variable

	if usedeviceCode {
		if prefix == "" {
			prefix = "AZURE_"
		}

		// TODO replace with constant appId
		applicationID := os.Getenv(fmt.Sprintf("%sPORTER_PLUGIN_APP_ID", prefix))
		env, err := getAzureEnvironment()
		if err != nil {
			return nil, errors.Wrap(err, "Failed to get Azure Environment")
		}
		deviceFlowConfig := azureauth.DeviceFlowConfig{
			TenantID:    tenantID,
			ClientID:    applicationID,
			Resource:    resourceID,
			AADEndpoint: env.ActiveDirectoryEndpoint,
		}
		authorizer, err = deviceFlowConfig.Authorizer()
		if err != nil {
			return nil, errors.Wrap(err, "Failed to create an azure authorizer from device flow")
		}
		return authorizer, nil
	}

	// NewAuthorizierFromEnvironment attempts to authenticate using credentials, certicates, user name and password and MSI however if we get here MSI login wll be skipped as env vars are set so one of the other methods will be attempted

	authorizer, err = auth.NewAuthorizerFromEnvironment()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create an azure authorizer from environment")
	}

	return authorizer, nil
}

// getAzureKeyVaultResourceID gets the Key Vault endpoint resource
func getAzureKeyVaultResourceID() (string, error) {

	resource := os.Getenv("AZURE_KEYVAULT_RESOURCE")
	if resource == "" {
		env, err := getAzureEnvironment()
		if err != nil {
			return "", err
		}
		resource = strings.TrimSuffix(env.KeyVaultEndpoint, "/")
	}

	return resource, nil
}

// getAzureEnvironment gets the Azure environment settings
func getAzureEnvironment() (*azure.Environment, error) {
	env := azure.PublicCloud
	var err error
	envName := os.Getenv("AZURE_ENVIRONMENT")
	if len(envName) > 0 {
		env, err = azure.EnvironmentFromName(envName)
		if err != nil {
			return nil, err
		}
	}

	return &env, nil
}

func noAzureAuthEnvVarsAreSet(azureAuthEnvVarNames []string) bool {
	for _, v := range azureAuthEnvVarNames {
		val := os.Getenv(v)
		if len(val) > 0 {
			return false
		}
	}
	return true
}

func validateOptions(usedeviceCode bool, useMSI bool, noVarsAreSet bool, prefix string) error {
	if prefix == "" {
		prefix = "AZURE_"
	}

	if usedeviceCode && useMSI {
		return errors.New("login-using-device-code amd login-using-msi should not be set at the same time")
	}

	if useMSI && !noVarsAreSet {
		return fmt.Errorf("%s* environment variables should not be set when trying to log in using MSI", prefix)
	}

	if usedeviceCode {
		tenantIDEnvVarName := prefix + strings.TrimPrefix(azureauth.TenantID, "AZURE_")
		var tenantID = os.Getenv(azureauth.TenantID)
		if len(tenantID) == 0 {
			return errors.New(fmt.Sprintf("login-using-device-code is set but %s is not set", tenantIDEnvVarName))
		}

		appIdVarName := fmt.Sprintf("%sPORTER_PLUGIN_APP_ID", prefix)
		applicationID := os.Getenv(appIdVarName)
		if len(applicationID) == 0 {
			return errors.New(fmt.Sprintf("login-using-device-code is set but %s is not set", appIdVarName))
		}
	}

	return nil
}
