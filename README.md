# Azure Plugins for Porter

This is a set of Azure plugins for [Porter](https://github.com/deislabs/porter).
 
[![Build Status](https://dev.azure.com/deislabs/porter/_apis/build/status/porter-azure-plugins?branchName=main)](https://dev.azure.com/deislabs/porter/_build/latest?definitionId=26&branchName=main)

## Install

The plugin is distributed as a single binary, `azure`. The following snippet will clone this repository, build the binary
and install it to **~/.porter/plugins/**.

```
go get get.porter.sh/plugin/azure/cmd/azure
cd $(go env GOPATH)/src/get.porter.sh/plugin/azure
make build install
```

After installing the plugin, you must modify your porter configuration file and select which plugin you want to use.

## Storage

Storage plugins allow Porter to store data, such as claims, parameters and credentials, in Azure's cloud.

### Blob

The `azure.blob` plugin stores data in Azure Blob Storage. 

1. Open, or create, `~/.porter/config.toml`.
1. Add the following line to activate the Azure blob storage plugin:

    ```toml
    default-storage-plugin = "azure.blob"
    ```

1. [Create a storage account][account]
1. [Create a container][container] named `porter`.
1. [Copy the connection string][connstring] for the storage account. Then set it as an environment variable named 
    `AZURE_STORAGE_CONNECTION_STRING`.

## Secrets

Secrets plugins allow Porter to inject secrets into credential or parameter sets.

For example, if your team has a shared key vault with a database password, you
can use the keyvault plugin to inject it as a credential or parameter when you install a bundle.

### Key Vault

The `azure.keyvault` plugin resolves credentials or parameters against secrets in Azure Key Vault.

1. Open, or create, `~/.porter/config.toml`
1. Add the following lines to activate the Azure keyvault secrets plugin:

    ```toml
    default-secrets = "mysecrets"
    
    [[secrets]]
    name = "mysecrets"
    plugin = "azure.keyvault"
    
    [secrets.config]
    vault = "myvault"
    ```
1. [Create a key vault][keyvault] and set the vault name in the config with name of the vault.

#### Authentication

Authentication to Azure can use any of the following methods, whichever mechanism is used the prinicpal that used to access key vault needs to be granted at [Get and List secret permission][keyvaultacl] on the vault 

1. Azure CLI 

1. [Create a service principal][sp] and create an Access Policy on the vault giving Get and List secret permissions.
1. Using credentials for the service principal set the environment variables: `AZURE_TENANT_ID`,`AZURE_CLIENT_ID`,  and `AZURE_CLIENT_SECRET`.

[account]: https://docs.microsoft.com/en-us/azure/storage/common/storage-quickstart-create-account?tabs=azure-portal
[container]: https://docs.microsoft.com/en-us/azure/storage/blobs/storage-quickstart-blobs-portal#create-a-container
[connstring]: https://docs.microsoft.com/en-us/azure/storage/common/storage-configure-connection-string?toc=%2fazure%2fstorage%2fblobs%2ftoc.json#view-and-copy-a-connection-string
[keyvault]: https://docs.microsoft.com/en-us/azure/key-vault/quick-create-portal#create-a-vault
[sp]: https://docs.microsoft.com/en-us/azure/active-directory/develop/howto-create-service-principal-portal
[keyvaultacl]: https://docs.microsoft.com/en-us/azure/key-vault/secrets/about-secrets#secret-access-control