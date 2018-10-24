package configurator

import (
        "os"
	"github.com/pkg/errors"
)

type loader interface {
	load(config interface{}) (error)
}

// Configurator is Configurator
type Configurator struct {
	loader     loader
}

// Load is load config
func (c *Configurator) Load() (*Config, error) {
        config := new(Config)
	err := c.loader.load(config)
        return config, err
}

func validateConfigFile(configFile string) (error) {
        _, err := os.Stat(configFile)
        if err != nil {
            return errors.Wrapf(err, "not exists config file (%v)", configFile)
        }
        f, err := os.Open(configFile) 
        defer f.Close()
        if err != nil {
            return errors.Wrapf(err, "can not open config file (%v)", configFile)
        }
        return nil
}

// NewConfigurator is create new configurator
func NewConfigurator(configFile string) (*Configurator, error) {
	err := validateConfigFile(configFile)
	if (err != nil) {
		return nil, errors.Wrapf(err, "invalid config file (%v)", configFile)
	}

	loader, err := newFileLoader(configFile)
	if (err != nil) {
		return nil, errors.Wrap(err, "can not create new file loader")
	}

	newConfigurator := &Configurator{
             loader: loader,
	}
	return newConfigurator, nil
}
