package agent

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Broker             string
	ClientID           string
	Username           string
	Password           string
	Prefix             string
	DisableRulesUpdate bool
}

type ConfigFile struct {
	Config     Config
	Parameters map[string]Parameter
	Rules      map[string]map[string]Rule
}

func ConfigFromFile(path string) (configFile *ConfigFile, e error) {
	data, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}
	var c ConfigFile
	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
