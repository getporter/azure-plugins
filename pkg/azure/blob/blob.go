package blob

import (
	"os"

	"github.com/deislabs/cnab-go/claim"
	"github.com/deislabs/porter-azure-plugins/pkg/azure/credentials"
	instancestorage "github.com/deislabs/porter/pkg/instance-storage"
	"github.com/deislabs/porter/pkg/instance-storage/claimstore"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

const Key = "instance-storage.azure-blob"

var _ instancestorage.ClaimStore = &Plugin{}

type Plugin struct {
	logger hclog.Logger
}

func NewPlugin() plugin.Plugin {
	// Create an hclog.Logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   Key,
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	return &claimstore.Plugin{
		Impl: &Plugin{
			logger: logger,
		},
	}
}

func (p Plugin) init() (claim.Store, error) {
	creds, err := credentials.GetCredentials()
	if err != nil {
		return claim.Store{}, err
	}

	crud := Store{
		Container:     "porter",
		CredentialSet: creds,
	}

	store := claim.NewClaimStore(crud)
	return store, nil
}

func (p *Plugin) List() ([]string, error) {
	store, err := p.init()
	if err != nil {
		return nil, err
	}

	return store.List()
}

func (p *Plugin) Store(c claim.Claim) error {
	store, err := p.init()
	if err != nil {
		return err
	}

	return store.Store(c)
}

func (p *Plugin) Read(name string) (*claim.Claim, error) {
	store, err := p.init()
	if err != nil {
		return nil, err
	}

	c, err := store.Read(name)
	return &c, err
}

func (p *Plugin) ReadAll() ([]claim.Claim, error) {
	store, err := p.init()
	if err != nil {
		return nil, err
	}

	return store.ReadAll()
}

func (p *Plugin) Delete(name string) error {
	store, err := p.init()
	if err != nil {
		return err
	}

	return store.Delete(name)
}
