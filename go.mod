module get.porter.sh/plugin/azure

go 1.13

require (
	get.porter.sh/porter v0.22.2-beta.1
	github.com/Azure/azure-pipeline-go v0.2.2
	github.com/Azure/azure-storage-blob-go v0.8.0
	github.com/cnabio/cnab-go v0.8.2-beta1
	github.com/hashicorp/go-hclog v0.9.2
	github.com/hashicorp/go-plugin v1.0.1
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d // indirect
	github.com/mattn/go-ieproxy v0.0.0-20190805055040-f9202b1cfdeb // indirect
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.4.0
	golang.org/x/net v0.0.0-20191007182048-72f939374954 // indirect
	google.golang.org/genproto v0.0.0-20191007204434-a023cd5227bd // indirect
	google.golang.org/grpc v1.24.0 // indirect
)

replace github.com/hashicorp/go-plugin => github.com/carolynvs/go-plugin v1.0.1-acceptstdin

replace get.porter.sh/porter => github.com/carolynvs/porter v0.8.0-ralpha.1.0.20200130141230-9eec42da929d

replace github.com/cnabio/cnab-go => github.com/carolynvs/cnab-go v0.0.0-20200129213214-320e82d9048c
