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
	get.porter.sh/porter v0.27.3-0.20200727164955-cad03081f589
	github.com/Azure/azure-pipeline-go v0.2.2
	github.com/Azure/azure-sdk-for-go v19.1.1+incompatible
	github.com/Azure/azure-storage-blob-go v0.8.0
	github.com/Azure/go-autorest/autorest v0.9.3
	github.com/Azure/go-autorest/autorest/azure/auth v0.4.2
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/cnabio/cnab-go v0.13.0-beta1
	github.com/hashicorp/go-hclog v0.9.2
	github.com/hashicorp/go-multierror v1.1.0
	github.com/hashicorp/go-plugin v1.0.1
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d // indirect
	github.com/mattn/go-ieproxy v0.0.0-20190805055040-f9202b1cfdeb // indirect
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v0.0.6
	github.com/stretchr/testify v1.5.1
	google.golang.org/genproto v0.0.0-20191007204434-a023cd5227bd // indirect
	google.golang.org/grpc v1.24.0 // indirect
)
