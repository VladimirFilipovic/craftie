package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vlad/craftie/internal/pkg"
	"gopkg.in/yaml.v3"
)

const (
	SessionSyncTime = time.Minute * 1
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

		if err := createConfigFile(cfgPath, defaultConfig()); err != nil {
			return nil, fmt.Errorf("failed to create default config file: %w", err)
		}
		fmt.Printf("Default config file created at %s\n", cfgPath)
	}

	// Load config file
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Expand tilde paths in config
	if err := config.expandPaths(); err != nil {
		return nil, fmt.Errorf("failed to expand paths: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
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

type Config struct {
	GoogleSheets  GoogleSheetsConfig `yaml:"google_sheets" mapstructure:"google_sheets"`
	Notifications NotificationConfig `yaml:"notifications" mapstructure:"notifications"`
	Logging       LoggingConfig      `yaml:"logging" mapstructure:"logging"`
	CSV           CSVConfig          `yaml:"csv" mapstructure:"csv"`
}

type GoogleSheetsConfig struct {
	SpreadsheetID     string `yaml:"spreadsheet_id" mapstructure:"spreadsheet_id"`
	SheetName         string `yaml:"sheet_name" mapstructure:"sheet_name"`
	CredentialsHelper string `yaml:"credentials_helper" mapstructure:"credentials_helper"`
	SyncInterval      time.Duration
	Enabled           bool `yaml:"enabled" mapstructure:"enabled"`
}

type NotificationConfig struct {
	Enabled          bool          `yaml:"enabled" mapstructure:"enabled"`
	ReminderInterval time.Duration `yaml:"reminder_interval" mapstructure:"reminder_interval"`
	SoundEnabled     bool          `yaml:"sound_enabled" mapstructure:"sound_enabled"`
}

type LoggingConfig struct {
	Level      string `yaml:"level" mapstructure:"level"`
	OutputFile string `yaml:"output_file" mapstructure:"output_file"`
}

// CSVConfig holds CSV file configuration
type CSVConfig struct {
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`
	FilePath string `yaml:"file_path" mapstructure:"file_path"`
}

func defaultConfig() *Config {
	return &Config{
		GoogleSheets: GoogleSheetsConfig{
			Enabled: false,
		},
		Notifications: NotificationConfig{
			Enabled:          true,
			ReminderInterval: 15 * time.Minute,
			SoundEnabled:     true,
		},
		Logging: LoggingConfig{
			Level: "info",
		},
		CSV: CSVConfig{
			Enabled:  false,
			FilePath: "",
		},
	}
}

// expandPaths expands tilde (~) in file paths to the user's home directory
func (c *Config) expandPaths() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	if strings.HasPrefix(c.GoogleSheets.CredentialsHelper, "~/") {
		c.GoogleSheets.CredentialsHelper = filepath.Join(homeDir, c.GoogleSheets.CredentialsHelper[2:])
	}

	if strings.HasPrefix(c.CSV.FilePath, "~/") {
		c.CSV.FilePath = filepath.Join(homeDir, c.CSV.FilePath[2:])
	}

	if strings.HasPrefix(c.Logging.OutputFile, "~/") {
		c.Logging.OutputFile = filepath.Join(homeDir, c.Logging.OutputFile[2:])
	}

	return nil
}

func (c *Config) Validate() error {
	if c.GoogleSheets.Enabled {
		// Check if credentials are available via helper or keyring
		if c.GoogleSheets.SpreadsheetID == "" {
			return pkg.NewValidationError("google_sheets.spreadsheet_id is required when Google Sheets is enabled")
		}

		if c.GoogleSheets.SheetName == "" {
			return pkg.NewValidationError("google_sheets.sheet_name is required when Google Sheets is enabled")
		}
	}

	if c.CSV.Enabled {
		if c.CSV.FilePath == "" {
			return pkg.NewValidationError("csv.file_path is required when CSV is enabled")
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
