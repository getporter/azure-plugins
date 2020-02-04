package azureconfig

type Config struct {

	// EnvConnectionString is the environment variable from which the connection
	// string should be loaded.
	EnvConnectionString string `json:"env"`

	// EnvServicePrincipalPrefix is the prefix applied to every service
	// principal environment variable For example, for a prefix of "DEV_AZURE_", the
	// variables would be "DEV_AZURE_TENANT_ID", "DEV_AZURE_CLIENT_ID",
	// "DEV_AZURE_CLIENT_SECRET". By default the prefix is "AZURE_".
	EnvServicePrincipalPrefix string `json:"env-sp-prefix"`

	// Vault is the name of the vault containing bundle secrets.
	Vault string `json:"vault"`
}
