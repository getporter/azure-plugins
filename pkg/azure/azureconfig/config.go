package azureconfig

type Config struct {

	// EnvConnectionString is the environment variable from which the connection
	// string should be loaded.
	EnvConnectionString string `json:"env"`

	// StorageAccount contains the name of the storage account to be used by the Azure storage plugin, if the azure connection environment variable is not set and this proeprty and StorageAccountResourceGroup are populated and the user is logged in with the Azure CLI
	// the Storage Account Key will be looked up at runtime using the logged in users credentials
	StorageAccount string `json:"account"`
	// StorageAccountResourceGroup contains the name of the resource group containing the storage account to be used by the Azure storage plugin, if the azure connection environment variable is not set and this property and StorageAccount are populated and the user is logged in with the Azure CLI
	// the Storage Account Key will be looked up at runtime using the logged in users credentials
	StorageAccountResourceGroup string `json:"resource-group"`
	// StorageAccountSubscriptionId contains the subscription id of the subscription to be used when looking up the Storage Account Key, if this is not set then the current CLI subscription will be used
	StorageAccountSubscriptionId string `json:"subscription-id"`

	// If set to true data will be compressed before being written to Table storage.
	StorageCompressData bool `json:"compress-data"`

	// Vault is the name of the vault containing bundle secrets.
	Vault string `json:"vault"`
	// VaultUrl is the full url of the vault containing bundle secrets.
	VaultUrl string `json:"vault-url"`
}
