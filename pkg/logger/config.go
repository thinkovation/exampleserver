package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type LogConfig struct {
	LogFile     string `yaml:"log_file"`
	LogToStdout bool   `yaml:"log_to_stdout"`
	Debug       bool   `yaml:"debug"`
	Rotation    struct {
		MaxSize    int  `yaml:"max_size"`    // maximum size in megabytes before rotating
		MaxAge     int  `yaml:"max_age"`     // maximum number of days to retain old log files
		MaxBackups int  `yaml:"max_backups"` // maximum number of old log files to retain
		Compress   bool `yaml:"compress"`    // compress rotated files
	} `yaml:"rotation"`
	Webhooks []WebhookConfig `yaml:"webhooks"`
}

type WebhookConfig struct {
	URL    string    `yaml:"url"`
	APIKey string    `yaml:"api_key"`
	Filter LogFilter `yaml:"filter"`
}

// DefaultConfig returns the default logging configuration
func DefaultConfig() *LogConfig {
	config := &LogConfig{
		LogFile:     "logs/app.log",
		LogToStdout: true,
		Debug:       false,
	}

	// Default rotation settings
	config.Rotation.MaxSize = 10    // 10 MB
	config.Rotation.MaxAge = 30     // 30 days
	config.Rotation.MaxBackups = 5  // Keep 5 old files
	config.Rotation.Compress = true // Compress old files

	return config
}

// LoadConfig loads the logger configuration from yaml file
func LoadConfig(configPath string) (*LogConfig, error) {
	// If no config file exists, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	config := &LogConfig{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Ensure log directory exists
	logDir := filepath.Dir(config.LogFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating log directory: %w", err)
	}

	return config, nil
}

// Modify Initialize to handle webhooks
func Initialize(configPath string) error {
	var err error
	once.Do(func() {
		// Load configuration
		var config *LogConfig
		config, err = LoadConfig(configPath)
		if err != nil {
			return
		}

		// Create logger instance
		defaultLogger, err = New(config)
		if err != nil {
			return
		}

		// Initialize webhooks if configured
		for _, webhookConfig := range config.Webhooks {
			fmt.Println("Initializing webhook plugin")
			if webhookConfig.URL == "" {
				fmt.Println("Webhook URL is empty - skipping")
				continue
			}
			webhook := NewWebhookPlugin(
				webhookConfig.URL,
				webhookConfig.APIKey,
				webhookConfig.Filter,
			)
			if err = defaultLogger.AddPlugin(webhook); err != nil {
				defaultLogger.Error("Failed to initialize webhook plugin: %v", err)
			}
		}
	})
	return err
}
