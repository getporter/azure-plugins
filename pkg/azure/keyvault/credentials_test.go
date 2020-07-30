package keyvault

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"get.porter.sh/plugin/azure/pkg/azure/azureconfig"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

func TestGet_GetCredentials(t *testing.T) {
	testcases := []struct {
		name         string
		config       *azureconfig.Config
		envVarsToSet map[string]string
		testfunc     func(t *testing.T, envVarsToSet map[string]string, config azureconfig.Config, logger hclog.Logger)
	}{
		{
			"GetCredentials using SPN",
			nil,
			map[string]string{
				"CLIENT_ID":     "CLIENT_ID_T1",
				"CLIENT_SECRET": "CLIENT_SECRET_T1",
				"TENANT_ID":     "TENANT_ID_T1"},
			validateSPNLogin,
		},
		{
			"GetCredentials using certificates",
			nil,
			map[string]string{
				"CERTIFICATE_PATH":     "./testdata/porter.pfx",
				"CERTIFICATE_PASSWORD": "password",
				"CLIENT_ID":            "CLIENT_ID_T4",
				"TENANT_ID":            "TENANT_ID_T4"},
			validateCertificateLogin,
		},
		{
			"GetCredentials using username and password",
			nil,
			map[string]string{
				"USERNAME":  "user",
				"PASSWORD":  "password",
				"CLIENT_ID": "CLIENT_ID_T7",
				"TENANT_ID": "TENANT_ID_T7"},
			validateUserNameAndPasswordLogin,
		},
		{
			"GetCredentials using MSI",
			&azureconfig.Config{
				LoginWithMSI: "true",
			},
			map[string]string{},
			validateMSILogin,
		},
		{
			"GetCredentials using az cli",
			nil,
			map[string]string{},
			validateazCLILogin,
		},
		{
			"GetCredentials using device code ",
			&azureconfig.Config{
				LoginWithDeviceCode: "true",
			},
			map[string]string{
				"TENANT_ID":            "TENANT_ID_T10",
				"PORTER_PLUGIN_APP_ID": "PORTER_APP_ID",
			},
			validateDeviceCodeLogin,
		},
		{
			"MSI Auth should have no AZURE vars",
			&azureconfig.Config{
				LoginWithMSI: "true",
			},
			map[string]string{
				"TENANT_ID": "TENANT_ID_T13"},
			validateMSILoginHasNoEnvVars,
		},
		{
			"Device Code needs TENANT_ID EnvVar",
			&azureconfig.Config{
				LoginWithDeviceCode: "true",
			},
			map[string]string{
				"PORTER_PLUGIN_APP_ID": "PORTER_APP_ID"},
			validateDeviceCodeLoginHasTenantId,
		},
		{
			"Device Code needs PORTER_APP_ID EnvVar",
			&azureconfig.Config{
				LoginWithDeviceCode: "true",
			},
			map[string]string{
				"TENANT_ID": "TENANT_ID_T12"},
			validateDeviceCodeLoginHasPorterAppId,
		},
	}
	env := os.Environ()
	for _, tc := range testcases {

		for _, prefix := range []string{
			"",
			"AZURE_",
			"DEV_",
		} {

			t.Run(fmt.Sprintf("%s with prefix: '%s'", tc.name, prefix), func(t *testing.T) {

				if tc.config == nil {
					tc.config = &azureconfig.Config{}
				}
				tc.config.EnvAzurePrefix = prefix

				logger := hclog.New(&hclog.LoggerOptions{
					Name:   strings.ReplaceAll(tc.name, " ", "_"),
					Output: os.Stderr,
					Level:  hclog.Error,
				})

				for k, v := range tc.envVarsToSet {
					var envVarName string
					var envVarValue string
					if len(prefix) == 0 {
						prefix = "AZURE_"
					}

					envVarName = fmt.Sprintf("%s%s", prefix, k)
					envVarValue = fmt.Sprintf("%s%s", prefix, v)
					if k == "CERTIFICATE_PATH" || k == "CERTIFICATE_PASSWORD" {
						envVarValue = v
					}
					os.Setenv(envVarName, envVarValue)
					tc.envVarsToSet[k] = envVarValue
				}

				tc.testfunc(t, tc.envVarsToSet, *tc.config, logger)
				resetEnvironmentVars(t, env)
			})
		}
	}
}

func resetEnvironmentVars(t *testing.T, env []string) {
	os.Clearenv()
	for _, e := range env {
		pair := strings.Split(e, "=")
		t.Logf("Resetting Env Variable: %s", pair[0])
		os.Setenv(pair[0], pair[1])
	}
}

func validateSPNLogin(t *testing.T, envVarsToSet map[string]string, config azureconfig.Config, logger hclog.Logger) {
	authorizer, err := GetCredentials(config, logger)
	assert.NoError(t, err)
	servicePrincipalToken := getServicePrincipalToken(t, authorizer)
	innerToken := getValue(t, reflect.ValueOf(servicePrincipalToken).Elem(), "inner")
	clientId := getValue(t, innerToken, "ClientID")
	assert.Equal(t, envVarsToSet["CLIENT_ID"], clientId.String())
	secretValue := getPrivateValue(t, innerToken, "Secret")
	secret, ok := secretValue.Interface().(*adal.ServicePrincipalTokenSecret)
	assert.True(t, ok)
	assert.Equal(t, envVarsToSet["CLIENT_SECRET"], secret.ClientSecret)
	oauthConfigValue := getPrivateValue(t, innerToken, "OauthConfig")
	oAuthConfig, ok := oauthConfigValue.Interface().(adal.OAuthConfig)
	assert.True(t, ok)
	assert.Contains(t, oAuthConfig.AuthorizeEndpoint.Path, envVarsToSet["TENANT_ID"])
}

func validateCertificateLogin(t *testing.T, envVarsToSet map[string]string, config azureconfig.Config, logger hclog.Logger) {
	authorizer, err := GetCredentials(config, logger)
	assert.NoError(t, err)
	servicePrincipalToken := getServicePrincipalToken(t, authorizer)
	innerToken := getValue(t, reflect.ValueOf(servicePrincipalToken).Elem(), "inner")
	clientId := getValue(t, innerToken, "ClientID")
	assert.Equal(t, envVarsToSet["CLIENT_ID"], clientId.String())
	secretValue := getPrivateValue(t, innerToken, "Secret")
	secret, ok := secretValue.Interface().(*adal.ServicePrincipalCertificateSecret)
	assert.True(t, ok)
	assert.NotZero(t, secret.Certificate)
	assert.NotZero(t, secret.PrivateKey)
	oauthConfigValue := getPrivateValue(t, innerToken, "OauthConfig")
	oAuthConfig, ok := oauthConfigValue.Interface().(adal.OAuthConfig)
	assert.True(t, ok)
	assert.Contains(t, oAuthConfig.AuthorizeEndpoint.Path, envVarsToSet["TENANT_ID"])
}

func validateUserNameAndPasswordLogin(t *testing.T, envVarsToSet map[string]string, config azureconfig.Config, logger hclog.Logger) {
	authorizer, err := GetCredentials(config, logger)
	assert.NoError(t, err)
	servicePrincipalToken := getServicePrincipalToken(t, authorizer)
	innerToken := getValue(t, reflect.ValueOf(servicePrincipalToken).Elem(), "inner")
	clientId := getValue(t, innerToken, "ClientID")
	assert.Equal(t, envVarsToSet["CLIENT_ID"], clientId.String())
	secretValue := getPrivateValue(t, innerToken, "Secret")
	secret, ok := secretValue.Interface().(*adal.ServicePrincipalUsernamePasswordSecret)
	assert.True(t, ok)
	assert.Equal(t, envVarsToSet["USERNAME"], secret.Username)
	assert.Equal(t, envVarsToSet["PASSWORD"], secret.Password)
	oauthConfigValue := getPrivateValue(t, innerToken, "OauthConfig")
	oAuthConfig, ok := oauthConfigValue.Interface().(adal.OAuthConfig)
	assert.True(t, ok)
	assert.Contains(t, oAuthConfig.AuthorizeEndpoint.Path, envVarsToSet["TENANT_ID"])
}

func validateMSILogin(t *testing.T, envVarsToSet map[string]string, config azureconfig.Config, logger hclog.Logger) {
	authorizer, err := GetCredentials(config, logger)
	assert.NoError(t, err)
	servicePrincipalToken := getServicePrincipalToken(t, authorizer)
	innerToken := getValue(t, reflect.ValueOf(servicePrincipalToken).Elem(), "inner")
	secretValue := getPrivateValue(t, innerToken, "Secret")
	_, ok := secretValue.Interface().(*adal.ServicePrincipalMSISecret)
	assert.True(t, ok)
}

func validateazCLILogin(t *testing.T, envVarsToSet map[string]string, config azureconfig.Config, logger hclog.Logger) {
	authorizer, err := GetCredentials(config, logger)
	if err != nil {
		// az cli errors can occur for a number of reasons, e.g. user is not logged in, az cli is not installed, az cli is installed but in non standard path etc.
		assert.True(t, strings.HasPrefix(err.Error(), "Failed to create an azure authorizer from azure cli: Invoking Azure CLI failed with the following error:"))
		return
	}
	getToken(t, authorizer)
}

func validateDeviceCodeLogin(t *testing.T, envVarsToSet map[string]string, config azureconfig.Config, logger hclog.Logger) {
	_, err := GetCredentials(config, logger)
	assert.NotNil(t, err)
	t.Log(t, err.Error())
	assert.True(t, strings.HasPrefix(err.Error(), "Failed to create an azure authorizer from device flow: failed to get oauth token from device flow: failed to start device auth flow: autorest/adal/devicetoken: Error occurred while handling response from the Device Endpoint: Error HTTP status"))
}

func validateMSILoginHasNoEnvVars(t *testing.T, envVarsToSet map[string]string, config azureconfig.Config, logger hclog.Logger) {
	_, err := GetCredentials(config, logger)
	CheckExpectedError(t, config.EnvAzurePrefix, "%s* environment variables should not be set when trying to log in using MSI", err)
}

func validateDeviceCodeLoginHasTenantId(t *testing.T, envVarsToSet map[string]string, config azureconfig.Config, logger hclog.Logger) {
	_, err := GetCredentials(config, logger)
	CheckExpectedError(t, config.EnvAzurePrefix, "login-using-device-code is set but %sTENANT_ID is not set", err)
}

func validateDeviceCodeLoginHasPorterAppId(t *testing.T, envVarsToSet map[string]string, config azureconfig.Config, logger hclog.Logger) {
	_, err := GetCredentials(config, logger)
	CheckExpectedError(t, config.EnvAzurePrefix, "login-using-device-code is set but %sPORTER_PLUGIN_APP_ID is not set", err)
}

func CheckExpectedError(t *testing.T, prefix string, fmtString string, err error) {
	if len(prefix) == 0 {
		prefix = "AZURE_"
	}

	assert.NotNil(t, err)
	t.Log(t, err.Error())
	assert.EqualError(t, err, fmt.Sprintf(fmtString, prefix))
}

func getServicePrincipalToken(t *testing.T, authorizer autorest.Authorizer) *adal.ServicePrincipalToken {
	assert.IsType(t, &autorest.BearerAuthorizer{}, authorizer)
	bearerAuthorizer := authorizer.(*autorest.BearerAuthorizer)
	assert.IsType(t, &adal.ServicePrincipalToken{}, bearerAuthorizer.TokenProvider())
	return bearerAuthorizer.TokenProvider().(*adal.ServicePrincipalToken)
}

func getToken(t *testing.T, authorizer autorest.Authorizer) *adal.Token {
	assert.IsType(t, &autorest.BearerAuthorizer{}, authorizer)
	bearerAuthorizer := authorizer.(*autorest.BearerAuthorizer)
	assert.IsType(t, &adal.Token{}, bearerAuthorizer.TokenProvider())
	return bearerAuthorizer.TokenProvider().(*adal.Token)
}

func getValue(t *testing.T, value reflect.Value, name string) reflect.Value {
	field := value.FieldByName(name)
	assert.False(t, field.IsZero())
	return field
}

func getPrivateValue(t *testing.T, value reflect.Value, name string) reflect.Value {
	field := getValue(t, value, name)
	field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
	assert.True(t, field.CanInterface())
	return field
}
