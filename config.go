package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type GuildConfig struct {
	Bots        []string `json:"bots"`
	Delete      bool     `json:"delete"`
	Warn        bool     `json:"warn"`
	Kick        bool     `json:"kick"`
	Ban         bool     `json:"ban"`
	Timeout     int      `json:"timeout"`
	WarnMessage string   `json:"warnMessage"`
}

type Config struct {
	ConfigVersion  int                    `json:"configVersionDontTouch"`
	Token          string                 `json:"token"`
	GlobalSettings GuildConfig            `json:"global"`
	GuildSettings  map[string]GuildConfig `json:"guildSettings"`
}

func (gc GuildConfig) Copy() GuildConfig {
	botsCopy := make([]string, len(gc.Bots))
	copy(botsCopy, gc.Bots)
	return GuildConfig{
		Bots: botsCopy, Delete: gc.Delete, Warn: gc.Warn, Kick: gc.Kick, Ban: gc.Ban,
		Timeout: gc.Timeout, WarnMessage: gc.WarnMessage,
	}
}

func (cfg *Config) ForGuildID(id string) GuildConfig {
	// Read Lock: allows multiple goroutines to read concurrently, but blocks writers.
	configMutex.RLock()
	defer configMutex.RUnlock()
	settings, exists := cfg.GuildSettings[id]
	if !exists {
		return cfg.GlobalSettings.Copy()
	}
	return settings
}

func LoadConfig() (*Config, error) {
	const CURRENT_CONFIG_VERSION = 1
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "."
	}

	file := filepath.Join(cfgPath, "config.json")
	config := &Config{
		ConfigVersion: CURRENT_CONFIG_VERSION,
		Token:         "YOUR_TOKEN_HERE",
		GlobalSettings: GuildConfig{
			// Viggle, DomoAI, Glifbot, InsightFaceSwap
			Bots:   []string{"1104973139257081909", "1153984868804468756", "1288638725869535283", "1090660574196674713"},
			Delete: true, Warn: true, Kick: false, Ban: false,
			Timeout:     0,
			WarnMessage: "Use of Discord's gen-AI tools is against our rules",
		},
		GuildSettings: make(map[string]GuildConfig),
	}

	data, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			if err := config.saveConfig(); err != nil {
				return nil, err
			} else {
				return config, fmt.Errorf("config file not found, created a blank sample with default settings")
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

	return config, nil
}

var saveChannel = make(chan struct{}, 1)

// Mutex to protect concurrent access to the Config struct's fields (like GuildSettings)
// RWMutex allows multiple readers or a single writer.
var configMutex sync.RWMutex

func (c *Config) saveConfig() error {
	// acquire a Read Lock: allows multiple goroutines to read concurrently, but blocks writers.
	configMutex.RLock()
	defer configMutex.RUnlock()

	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "."
	}
	if err := os.MkdirAll(cfgPath, 0770); err != nil {
		return err
	}

	file := filepath.Join(cfgPath, "config.json")

	if data, err := json.MarshalIndent(c, "", "  "); err == nil {
		return os.WriteFile(file, data, 0660)
	} else {
		return err
	}
}

func InitConfigSaver(c *Config) {
	go func() {
		for range saveChannel {
			if err := c.saveConfig(); err != nil {
				fmt.Println("Failed to save config:", err)
			}
		}
	}()
}

func (cfg *Config) UpdateGuildSettings(guildID string, settings GuildConfig) {
	// acquire a Write Lock: blocks both readers and other writers.
	configMutex.Lock()
	defer configMutex.Unlock()

	if cfg.GuildSettings == nil {
		cfg.GuildSettings = make(map[string]GuildConfig)
	}
	cfg.GuildSettings[guildID] = settings

	select {
	case saveChannel <- struct{}{}:
		// signal sent successfully
	default:
		// channel full, signal dropped (coalescing)
	}
}
