package azureconfig

type Config struct {
	// EnvAzurePrefix is the prefix applied to every azure
	// environment variable For example, for a prefix of "DEV_AZURE_", the
	// variables would be "DEV_AZURE_TENANT_ID", "DEV_AZURE_CLIENT_ID",
	// "DEV_AZURE_CLIENT_SECRET". By default the prefix is "AZURE_".
	EnvAzurePrefix string `json:"env-azure-prefix"`

	// Vault is the name of the vault containing bundle secrets.
	Vault string `json:"vault"`
	// VaultUrl is the full url of the vault containing bundle secrets.
	VaultUrl string `json:"vault-url"`
}
