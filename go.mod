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
	get.porter.sh/porter v0.28.1
	github.com/Azure/azure-pipeline-go v0.2.2
	github.com/Azure/azure-sdk-for-go v44.2.0+incompatible
	github.com/Azure/azure-storage-blob-go v0.8.0
	github.com/Azure/go-autorest/autorest v0.11.2
	github.com/Azure/go-autorest/autorest/adal v0.9.0
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.0
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.0
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.0 // indirect
	github.com/cnabio/cnab-go v0.13.4-0.20200817181428-9005c1da4354
	github.com/google/uuid v1.1.1
	github.com/hashicorp/go-hclog v0.9.2
	github.com/hashicorp/go-plugin v1.0.1
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d // indirect
	github.com/mattn/go-ieproxy v0.0.0-20190805055040-f9202b1cfdeb // indirect
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/spf13/cobra v0.0.6
	github.com/stretchr/testify v1.5.1
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	google.golang.org/genproto v0.0.0-20191007204434-a023cd5227bd // indirect
	google.golang.org/grpc v1.24.0 // indirect
)
