package config

import "gopkg.in/yaml.v3"
import "os"
import "fmt"

type Config struct {
	Telegram struct {
		Token        string `yaml:"token"`
		Enabled      bool   `yaml:"enabled"`
		NotifyChatId string `yaml:"notify_chat_id"`
	} `yaml:"telegram"`
	Users    []User `yaml:"users"`
	Hardware Gpio   `yaml:"hardware"`
	Door     struct {
		RelayOffHoldTimeS int `yaml:"relay_off_hold_time_s"`
	} `yaml:"door"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	return &cfg, nil
}
