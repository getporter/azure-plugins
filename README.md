# Azure Plugins for Porter

This is a set of Azure plugins for [Porter](https://github.com/getporter/porter).
 
[![Build Status](https://dev.azure.com/getporter/porter/_apis/build/status/azure-plugins?branchName=main)](https://dev.azure.com/getporter/porter/_build/latest?definitionId=8&branchName=main)

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

The `azure.blob` plugin stores data in Azure Blob Storage. The plugin requires a storage account name and storage account key. This can be provided as a connection string in an environment variable or can be looked up at run time if the user is logged in with the Azure CLI.

1. [Create a storage account][account]
1. [Create a container][container] named `porter`.
1. Open, or create, `~/.porter/config.toml`.

#### Use a connection string

1. Add the following line to activate the Azure blob storage plugin:

    ```toml
    default-storage-plugin = "azure.blob"
    ```
1. [Copy the connection string][connstring] for the storage account. Then set it as an environment variable named 
    `AZURE_STORAGE_CONNECTION_STRING`.

#### Use the Azure CLI

1. Add the following lines to activate the Azure blob storage plugin and configure storage account details:

    ```toml
    default-storage = "azureblob"

    [[storage]]
    name = "azureblob"
    plugin = "azure.blob"

    [storage.config]
    account="storage account name"
    resource-group="storage account resource group"

    ```

If the machine you are using is already logged in with the Azure CLI, then the same security context will be used to lookup the keys for the storage account. By default it will use the current subscription (the one returned by the command `az account show`). To set the subscription explicitly add the following line to the `[storage.config]`.

 ```toml
 subscription-id="storage account subscription id"
 ```

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

Authentication to Azure can use any of the following methods. Whichever mechanism is used, the principal that is used to access key vault needs to be granted at least [Get and List secret permissions][keyvaultacl] on the vault. However, if you authenticate using the Azure CLI and are logged in with the account that created the key vault in the portal then you will already have this permission.

1. **[Azure CLI][azurecli]**. - By default if the machine you are using is already logged in with the Azure CLI then the same security context will be used for the `azure.keyvault` plugin without any additional configuration.

1. **Use a service principal ([azure portal][sp] ) and an application secret ([azure portal][secret] or [azure cli][passwordcli])**. - Use the service principal details to set the environment variables `AZURE_TENANT_ID` and `AZURE_CLIENT_ID`. Then set the environment variable `AZURE_CLIENT_SECRET`using the application secret .

1. **Use a service principal ([azure portal][sp]) and a certificate ([azure portal][certificate]  or [azure cli][certcli])**. - Use the service principal details to set the environment variables `AZURE_TENANT_ID` and `AZURE_CLIENT_ID`. Then using the certificate file path and password set the environment variables `AZURE_CERTIFICATE_PATH` and `AZURE_CERTIFICATE_PASSWORD`.

1. **Username and Password** - Log in with user name and password.  Set the environment variables `AZURE_USERNAME` and `AZURE_PASSWORD`. This doesn't work with Microsoft accounts or accounts that have two-factor authentication enabled.

[account]: https://docs.microsoft.com/en-us/azure/storage/common/storage-quickstart-create-account?tabs=azure-portal
[container]: https://docs.microsoft.com/en-us/azure/storage/blobs/storage-quickstart-blobs-portal#create-a-container
[connstring]: https://docs.microsoft.com/en-us/azure/storage/common/storage-configure-connection-string?toc=%2fazure%2fstorage%2fblobs%2ftoc.json#view-and-copy-a-connection-string
[keyvault]: https://docs.microsoft.com/en-us/azure/key-vault/quick-create-portal#create-a-vault
[sp]: https://docs.microsoft.com/en-us/azure/active-directory/develop/howto-create-service-principal-portal
[keyvaultacl]: https://docs.microsoft.com/en-us/azure/key-vault/secrets/about-secrets#secret-access-control
[azurecli]: https://docs.microsoft.com/en-us/cli/azure/reference-index?view=azure-cli-latest#az-login
[secret]: https://docs.microsoft.com/en-us/azure/active-directory/develop/howto-create-service-principal-portal#create-a-new-application-secret
[certificate]: https://docs.microsoft.com/en-us/azure/active-directory/develop/howto-create-service-principal-portal#upload-a-certificate
[passwordcli]:https://docs.microsoft.com/en-us/cli/azure/create-an-azure-service-principal-azure-cli?view=azure-cli-latest#password-based-authentication
[certcli]:https://docs.microsoft.com/en-us/cli/azure/create-an-azure-service-principal-azure-cli?view=azure-cli-latest#certificate-based-authentication
