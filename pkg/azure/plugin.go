package azure

import (
	"bufio"
	"encoding/json"
	"io/ioutil"

	"get.porter.sh/plugin/azure/pkg/azure/config"
	"get.porter.sh/porter/pkg/context"
	"github.com/pkg/errors"
)

type Plugin struct {
	*context.Context
	config.Config
}

// New azure plugin client, initialized with useful defaults.
func New() *Plugin {
	return &Plugin{
		Context: context.New(),
	}
}

func (p *Plugin) LoadConfig() error {
	reader := bufio.NewReader(p.In)
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return errors.Wrap(err, "could not read stdin")
	}

	if len(b) == 0 {
		return nil
	}

	err = json.Unmarshal(b, &p.Config)
	if err != nil {
		return errors.Wrapf(err, "error unmarshaling stdin %q as azure.Config", string(b))
	}

	return nil
}
