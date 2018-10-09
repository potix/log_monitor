package configurator

import (
	"github.com/pkg/errors"
)

type loader interface {
	load(config interface{}) (error)
}

type Configurator struct {
	loader     loader
}

func (c *Configurator) Load(config interface{}) (error) {
	return c.loader.load(config)
}

func (c *Configurator) validateConfigFile() (error) {
        // XXX TODO check file
}

func NewConfigurator(configFile string) (*Configurator, error) {
	err := validateConfigFile(configFile)
	if (err != nil) {
		return nil, errors.Wrapf(err, "invalid config file (%v)", configFile)
	}

	loader, err = newFileLoader(configFile)
	if (err != nil) {
		return nil, errors.Wrap(err, "can not create new file loader")
	}

	newConfigurator := &Configurator{
             loader = loader
	}
	return newConfigurator, nil
}
