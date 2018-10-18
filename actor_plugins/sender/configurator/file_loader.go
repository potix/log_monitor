package configurator

import (
	"github.com/pkg/errors"
	"github.com/BurntSushi/toml"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"path/filepath"
	"io/ioutil"
	"os"
)

type fileLoader struct {
	configFile string
}

func (f *fileLoader) load(config interface{}) (error) {
	ext := filepath.Ext(f.configFile)
	switch ext {
	case ".tml":
		fallthrough
	case ".toml":
		_, err := toml.DecodeFile(f.configFile, config)
		if err != nil {
			return errors.Wrapf(err, "can not decode file with toml (%v)", f.configFile)
		}
	case ".yml":
		fallthrough
	case ".yaml":
		buf, err := ioutil.ReadFile(f.configFile)
		if err != nil {
			return errors.Wrapf(err, "can not read file with yaml (%v)", f.configFile)
		}
		err = yaml.Unmarshal(buf, config)
		if err != nil {
			return errors.Wrapf(err, "can not decode file with yaml (%v)", f.configFile)
		}
	case ".jsn":
		fallthrough
	case ".json":
		buf, err := ioutil.ReadFile(f.configFile)
		if err != nil {
			return errors.Wrapf(err,"can not read file with json (%v)", f.configFile)
		}
		err = json.Unmarshal(buf, config)
		if err != nil {
			return errors.Wrapf(err, "can not decode file with json (%v)", f.configFile)
		}
	default:
		return errors.Errorf("unexpected file extension (%v)", ext)
	}
	return nil
}

func newFileLoader(configFile string) (loader, error) {
	_, err := os.Stat(configFile)
	if err != nil {
		return nil, errors.Wrapf(err, "not exists config file (%v)", configFile)
	}
	return &fileLoader{
            configFile: configFile,
        }, nil
}
