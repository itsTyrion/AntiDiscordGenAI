package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Token  string   `json:"token"`
	Bots   []string `json:"bots"`
	Delete bool     `json:"delete"`
	Warn   bool     `json:"warn"`
	Kick   bool     `json:"kick"`
	Ban    bool     `json:"ban"`
}

func LoadConfig() (*Config, error) {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "."
	}

	file := filepath.Join(cfgPath, "config.json")
	data, err := os.ReadFile(file)
	var config = Config{
		Token:  "YOUR_TOKEN_HERE",
		Bots:   []string{"1153984868804468756", "1153984868804468756", "1288638725869535283", "1090660574196674713"},
		Delete: true, Warn: true, Kick: false, Ban: false,
	}
	if err != nil {
		if os.IsNotExist(err) {
			if err := config.SaveConfig(); err != nil {
				return nil, err
			} else {
				return &config, fmt.Errorf("config file not found, created a blank sample")
			}
		}
		return nil, err
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.Token == "" || config.Token == "YOUR_TOKEN_HERE" {
		return nil, fmt.Errorf("token is blank/unset")
	}

	return &config, nil
}

func (c *Config) SaveConfig() error {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "."
	}
	if err := os.MkdirAll(cfgPath, 0770); err != nil {
		return err
	}

	file := filepath.Join(cfgPath, "config.json")

	if data, err := json.MarshalIndent(c, "", "  "); err == nil {
		return os.WriteFile(file, data, 0640)
	} else {
		return err
	}
}
