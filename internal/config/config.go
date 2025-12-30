package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/vlad/craftie/internal/pkg"
	"gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from the specified path or creates default if it doesn't exist
func LoadConfig(cfgPath string) (*Config, error) {
	if cfgPath == "" {
		fmt.Println("No config path provided, using the default one")
		cfgPath = DefaultConfigPath()
	}

	var configFileExists bool = true
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		configFileExists = false
	}

	var isDefaultConfigPath = cfgPath == DefaultConfigPath()
	if !configFileExists && !isDefaultConfigPath {
		return nil, fmt.Errorf("config file doesn't exist")
	}

	if !configFileExists {
		fmt.Println("Config doesn't exist, generating default...")
		config := defaultConfig()

		if err := createConfigFile(cfgPath, config); err != nil {
			return nil, fmt.Errorf("failed to create default config file: %w", err)
		}
		fmt.Printf("Default config file created at %s\n", cfgPath)
	}

	// Load existing config file
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config *Config
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func createConfigFile(configPath string, config *Config) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create config file
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	configContent, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config file: %w", err)
	}

	_, err = file.Write(configContent)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// DefaultConfigPath returns the default configuration file path
func DefaultConfigPath() string {
	// Use XDG_CONFIG_HOME if set, otherwise default to ~/.config
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to system config
			return "/etc/craftie/craftie.yaml"
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(configDir, "craftie", "craftie.yaml")
}

// Config represents the application configuration
type Config struct {
	GoogleSheets  GoogleSheetsConfig `yaml:"google_sheets" mapstructure:"google_sheets"`
	Notifications NotificationConfig `yaml:"notifications" mapstructure:"notifications"`
	Logging       LoggingConfig      `yaml:"logging" mapstructure:"logging"`
}

// GoogleSheetsConfig holds Google Sheets API configuration
type GoogleSheetsConfig struct {
	CredentialsFile string `yaml:"credentials_file" mapstructure:"credentials_file"`
	SpreadsheetID   string `yaml:"spreadsheet_id" mapstructure:"spreadsheet_id"`
	SheetName       string `yaml:"sheet_name" mapstructure:"sheet_name"`
	SyncInterval    time.Duration
	Enabled         bool `yaml:"enabled" mapstructure:"enabled"`
}

// NotificationConfig holds notification system configuration
type NotificationConfig struct {
	Enabled          bool          `yaml:"enabled" mapstructure:"enabled"`
	ReminderInterval time.Duration `yaml:"reminder_interval" mapstructure:"reminder_interval"`
	SoundEnabled     bool          `yaml:"sound_enabled" mapstructure:"sound_enabled"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level" mapstructure:"level"`
	OutputFile string `yaml:"output_file" mapstructure:"output_file"`
}

func defaultConfig() *Config {
	return &Config{
		GoogleSheets: GoogleSheetsConfig{
			Enabled: false,
		},
		Notifications: NotificationConfig{
			Enabled:          true,
			ReminderInterval: 1 * time.Hour,
			SoundEnabled:     false,
		},
		Logging: LoggingConfig{
			Level: "info",
		},
	}
}

func (c *Config) Validate() error {
	if c.GoogleSheets.Enabled {
		if c.GoogleSheets.CredentialsFile == "" {
			return pkg.NewValidationError("google_sheets.credentials_file is required when Google Sheets is enabled")
		}

		// Check if credentials file exists
		if _, err := os.Stat(c.GoogleSheets.CredentialsFile); os.IsNotExist(err) {
			return pkg.NewValidationError("Google Sheets credentials file not found: " + c.GoogleSheets.CredentialsFile)
		}

		if c.GoogleSheets.SpreadsheetID == "" {
			return pkg.NewValidationError("google_sheets.spreadsheet_id is required when Google Sheets is enabled")
		}
	}

	validLevels := map[string]bool{
		"trace": true, "debug": true, "info": true,
		"warn": true, "error": true, "fatal": true, "panic": true,
	}
	if !validLevels[c.Logging.Level] {
		return pkg.NewValidationError("logging.level must be one of: trace, debug, info, warn, error, fatal, panic")
	}
	return nil
}
