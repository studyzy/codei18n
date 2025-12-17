package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the project configuration
type Config struct {
	SourceLanguage      string            `json:"sourceLanguage" mapstructure:"sourceLanguage"`
	LocalLanguage       string            `json:"localLanguage" mapstructure:"localLanguage"`
	ExcludePatterns     []string          `json:"excludePatterns" mapstructure:"excludePatterns"`
	TranslationProvider string            `json:"translationProvider" mapstructure:"translationProvider"`
	TranslationConfig   map[string]string `json:"translationConfig" mapstructure:"translationConfig"`
	BatchSize           int               `json:"batchSize" mapstructure:"batchSize"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		SourceLanguage:      "en",
		LocalLanguage:       "zh-CN",
		ExcludePatterns:     []string{".git/**", "vendor/**", ".codei18n/**"},
		TranslationProvider: "google",
		TranslationConfig:   make(map[string]string),
		BatchSize:           10,
	}
}

// LoadConfig loads the configuration from Viper into the Config struct
func LoadConfig() (*Config, error) {
	var cfg Config
	// Set defaults
	defaults := DefaultConfig()
	viper.SetDefault("sourceLanguage", defaults.SourceLanguage)
	viper.SetDefault("localLanguage", defaults.LocalLanguage)
	viper.SetDefault("excludePatterns", defaults.ExcludePatterns)
	viper.SetDefault("translationProvider", defaults.TranslationProvider)
	viper.SetDefault("batchSize", defaults.BatchSize)

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	// Ensure BatchSize is positive
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 10
	}
	return &cfg, nil
}

// SaveConfig writes the configuration to the .codei18n/config.json file
func SaveConfig(cfg *Config) error {
	// Create .codei18n directory if it doesn't exist
	configDir := ".codei18n"
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "config.json")
	file, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}
