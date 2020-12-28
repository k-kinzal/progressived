package conf

import (
	"bytes"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
	"io/ioutil"
	"os"
)

func Unmarshal(buf []byte) (*Config, error) {
	validate := validator.New()
	dec := yaml.NewDecoder(
		bytes.NewReader(buf),
		yaml.Validator(validate),
	)
	var config *Config
	if err := dec.Decode(&config); err != nil {
		return nil, fmt.Errorf(yaml.FormatError(err, false, true))
	}

	return config, nil
}

func Read(configPath string) (*Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("conf: %s: no such file or directory", configPath)
	}
	buf, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("conf: %s: %s", configPath, err)
	}
	conf, err := Unmarshal(buf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}