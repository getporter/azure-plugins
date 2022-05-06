module get.porter.sh/plugin/azure

go 1.13

replace (

	// This is a temporary reference to the porter's release/v1 branch that
	// conatins the new secret plugin protocol
	get.porter.sh/porter => get.porter.sh/porter v1.0.0-alpha.19.0.20220506213150-2201f7f910bc
	github.com/hashicorp/go-plugin => github.com/getporter/go-plugin v1.4.3-improved-configuration.1

	// Fixes https://github.com/spf13/viper/issues/761
	github.com/spf13/viper => github.com/getporter/viper v1.7.1-porter.2.0.20210514172839-3ea827168363

)

require (
	get.porter.sh/magefiles v0.2.2
	get.porter.sh/porter v0.0.0-00010101000000-000000000000
	github.com/Azure/azure-pipeline-go v0.2.2
	github.com/Azure/azure-sdk-for-go v44.2.0+incompatible
	github.com/Azure/azure-storage-blob-go v0.8.0
	github.com/Azure/go-autorest/autorest v0.11.18
	github.com/Azure/go-autorest/autorest/adal v0.9.13
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.0
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.0 // indirect
	github.com/cnabio/cnab-go v0.23.2
	github.com/hashicorp/go-hclog v0.14.1
	github.com/hashicorp/go-plugin v1.4.0
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.1
)
