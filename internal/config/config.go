package config

import (
	"github.com/kovetskiy/ko"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Listen string `yaml:"listen" env:"LISTEN" default:"127.0.0.1:3451"`
}

func Load(path string) (*Config, error) {
	config := &Config{}
	err := ko.Load(path, config, yaml.Unmarshal, ko.RequireFile(false))
	if err != nil {
		return nil, err
	}

	return config, nil
}
