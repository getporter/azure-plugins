# Azure Plugins for Porter

This is a set of Azure plugins for [Porter](https://github.com/deislabs/porter).
 
[![Build Status](https://dev.azure.com/deislabs/porter/_apis/build/status/porter-azure-plugins?branchName=master)](https://dev.azure.com/deislabs/porter/_build/latest?definitionId=26&branchName=master)

## Install

The plugins are distributed as a single binary, `azure`. The following snippet will clone this repository, build the binary
and install it to **~/.porter/plugins/**.

```
go get github.com/deislabs/porter-azure-plugins
cd $(go env GOPATH)/src/github.com/deislabs/porter-azure-plugins
make build install
```

After installing the plugin, you must modify your porter configuration file and select which plugin you want to use.

## Instance Storage

Instance Storage plugins allow Porter to store a record of installed bundles in a remote location.

### Blob

The `instance-storage.azure.blob` plugin stores records of your bundle instances in Azure Blob Storage. 

1. Open, or create, `~/.porter/config.toml`.
1. Add the following line to instruct Porter to use the azure blob storage plugin

    ```toml
    instance-storage-plugin = "azure.blob"
    ```

1. [Create a storage account][account]
1. [Create a container][container] named `porter`.
1. [Copy the connection string][connstring] for the storage account. Then set it as an environment variable named 
    `AZURE_STORAGE_CONNECTION_STRING`.

[account]: https://docs.microsoft.com/en-us/azure/storage/common/storage-quickstart-create-account?tabs=azure-portal
[container]: https://docs.microsoft.com/en-us/azure/storage/blobs/storage-quickstart-blobs-portal#create-a-container
[connstring]: https://docs.microsoft.com/en-us/azure/storage/common/storage-configure-connection-string?toc=%2fazure%2fstorage%2fblobs%2ftoc.json#view-and-copy-a-connection-string
