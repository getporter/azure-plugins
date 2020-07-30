package azureconfig

type Config struct {

	// EnvConnectionString is the environment variable from which the connection
	// string should be loaded.
	EnvConnectionString string `json:"env"`

	// EnvAzurePrefix is the prefix applied to every azure
	// environment variable For example, for a prefix of "DEV_AZURE_", the
	// variables would be "DEV_AZURE_TENANT_ID", "DEV_AZURE_CLIENT_ID",
	// "DEV_AZURE_CLIENT_SECRET". By default the prefix is "AZURE_".
	EnvAzurePrefix string `json:"env-azure-prefix"`

	// Vault is the name of the vault containing bundle secrets.
	Vault string `json:"vault"`

	// Enable DeviceCode login
	// Set to true to enable device code login
	// Device code login will only be used if AZURE_TENANT_ID variable is set
	LoginWithDeviceCode string `json:"login-using-device-code"`

	// Attempt MSI login
	// Set to true to enable msi login
	// MSI login will only be used if AZURE_* variables are not set
	// If msi login is not set and if no AZURE_* variables are  set then AZ cli login will be attempted
	LoginWithMSI string `json:"login-using-msi"`
}
