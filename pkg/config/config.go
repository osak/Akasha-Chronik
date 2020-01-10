package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	Twitter TwitterConfig `yaml:"twitter,omitempty"`
	Pixiv   PixivConfig   `yaml:"pixiv,omitempty"`
}

type TwitterConfig struct {
	ConsumerKey    string `yaml:"consumer_key"`
	ConsumerSecret string `yaml:"consumer_secret"`
	AccessToken    string `yaml:"access_token"`
	AccessSecret   string `yaml:"access_secret"`
}

type PixivConfig struct {
	PhpSessID string `yaml:"php_sess_id"`
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open file %s: %w", path, err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	config := Config{}
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("invalid YAML file %s: %w", path, err)
	}

	return &config, nil
}
