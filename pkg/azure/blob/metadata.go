package blob

import "github.com/cnabio/cnab-go/schema"

type Metadata struct {
	ClaimSchemaVersion schema.Version `json:"claim-schema-version"`
}
