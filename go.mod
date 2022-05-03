module get.porter.sh/plugin/azure

go 1.13

replace (
	// This is a temporary fork (branch: blob-indexes) that we are using until the azure sdk for go supports Blob Indexes (tags)
	github.com/Azure/azure-pipeline-go => github.com/carolynvs/azure-pipeline-go v0.2.3-0.20200624142537-02d87bc5483d

	// This is a temporary fork (branch: blob-indexes) that we are using until the azure sdk for go supports Blob Indexes (tags)
	github.com/Azure/azure-storage-blob-go => github.com/carolynvs/azure-storage-blob-go v0.9.1-0.20200622143949-348533d1f045

	github.com/hashicorp/go-plugin => github.com/carolynvs/go-plugin v1.0.1-acceptstdin
)

require (
	get.porter.sh/porter v1.0.0-alpha.19.0.20220502130939-4a3c3af95042
	github.com/Azure/azure-pipeline-go v0.2.2
	github.com/Azure/azure-sdk-for-go v44.2.0+incompatible
	github.com/Azure/azure-storage-blob-go v0.8.0
	github.com/Azure/go-autorest/autorest v0.11.18
	github.com/Azure/go-autorest/autorest/adal v0.9.13
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.0
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.0
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.0 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/bitly/go-hostpool v0.0.0-20171023180738-a3a6125de932 // indirect
	github.com/cnabio/cnab-go v0.23.1
	github.com/dnaeon/go-vcr v1.1.0
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8 // indirect
	github.com/gobuffalo/packr/v2 v2.8.0 // indirect
	github.com/godbus/dbus v4.1.0+incompatible // indirect
	github.com/google/uuid v1.3.0
	github.com/hashicorp/go-hclog v0.14.1
	github.com/hashicorp/go-plugin v1.4.0
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d // indirect
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.1
	github.com/xlab/handysort v0.0.0-20150421192137-fb3537ed64a1 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	vbom.ml/util v0.0.0-20180919145318-efcd4e0f9787 // indirect
)
