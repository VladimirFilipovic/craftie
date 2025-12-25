package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/vlad/craftie/internal/pkg"
	"gopkg.in/yaml.v3"
)

type ConfigManager struct {
	Config *Config
	viper  *viper.Viper
}

func NewConfigManager() *ConfigManager {
	viper := viper.New()

	return &ConfigManager{
		viper: viper,
	}
}

func (cm *ConfigManager) LoadConfig(cfgPath string) (*Config, error) {
	var config *Config
	var saveDefaultConfig = false

	// if no path provided use default one and generate it it doesn't already exists
	if cfgPath == "" {
		fmt.Println("No config path provided using the default one")

		defaultCfg, err := pkg.GetExpandedPathWithHome(DefaultConfigPath())
		cfgPath = defaultCfg

		if err != nil {
			return nil, err
		}

		// if default config isn't doesn't exist as file, create it
		if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
			fmt.Println("default config  doesn't exist generating it..")
			saveDefaultConfig = true
		}

	}

	// load existing config file
	if !saveDefaultConfig {
		fmt.Println(cfgPath)
		fmt.Println(filepath.Dir(cfgPath))
		fmt.Println(filepath.Base(cfgPath))

		cm.viper.AddConfigPath(filepath.Dir(cfgPath))
		cm.viper.SetConfigName(filepath.Base(cfgPath))
		cm.viper.SetConfigType("yaml")
		cm.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		cm.viper.AutomaticEnv()
		if err := cm.viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		config = defaultConfig()
		if err := cm.viper.Unmarshal(config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	if saveDefaultConfig {
		config = defaultConfig()
		if err := createConfigFile(cfgPath, config); err != nil {
			return nil, fmt.Errorf("failed to create default config file: %w", err)
		}
		fmt.Printf("Default config file created at %s\n", cfgPath)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func (m *ConfigManager) Set(key string, value interface{}) {
	m.viper.Set(key, value)
	// Re-unmarshal to update the config struct
	m.viper.Unmarshal(m.Config)
}

func (m *ConfigManager) Get(key string) interface{} {
	return m.viper.Get(key)
}

func expandPaths(conf *Config) error {
	var err error
	if conf.GoogleSheets.Enabled {
		expandedPath, err := pkg.GetExpandedPathWithHome(conf.GoogleSheets.CredentialsFile)
		if err != nil {
			return err
		}
		conf.GoogleSheets.CredentialsFile = expandedPath
	}

	expandedPath, err := pkg.GetExpandedPathWithHome(conf.Storage.DatabasePath)
	if err != nil {
		return err
	}
	conf.Storage.DatabasePath = expandedPath

	if conf.DaemonSocketPath != "" {
		expandedPath, err = pkg.GetExpandedPathWithHome(conf.DaemonSocketPath)
		if err != nil {
			return err
		}

		conf.DaemonSocketPath = expandedPath
	}

	expandedPath, err = pkg.GetExpandedPathWithHome(conf.Logging.OutputFile)
	if err != nil {
		return err
	}
	conf.Logging.OutputFile = expandedPath

	return nil
}

func createConfigFile(configPath string, config *Config) error {
	// Ensure directory exists, should be done by the installer
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
		return fmt.Errorf("creating config file issue: failed to marshal config file")
	}

	_, err = file.Write(configContent)
	if err != nil {
		return fmt.Errorf("creating config file issue: failed to write to a file")
	}

	return nil
}

// TODO: Add support for other OS-es
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
	DaemonSocketPath string             `yaml:"socket_path" mapstructure:"socket_path"`
	GoogleSheets     GoogleSheetsConfig `yaml:"google_sheets" mapstructure:"google_sheets"`
	Notifications    NotificationConfig `yaml:"notifications" mapstructure:"notifications"`
	Storage          StorageConfig      `yaml:"storage" mapstructure:"storage"`
	Logging          LoggingConfig      `yaml:"logging" mapstructure:"logging"`
}

// GoogleSheetsConfig holds Google Sheets API configuration
type GoogleSheetsConfig struct {
	CredentialsFile string `yaml:"credentials_file" mapstructure:"credentials_file"`
	SpreadsheetID   string `yaml:"spreadsheet_id" mapstructure:"spreadsheet_id"`
	SheetName       string `yaml:"sheet_name" mapstructure:"sheet_name"`
	SyncInterval    time.Duration
	Enabled         bool `yaml:"enabled" mapstructure:"enabled"`
	RetryAttempts   int
	RetryDelay      time.Duration
}

// NotificationConfig holds notification system configuration
type NotificationConfig struct {
	Enabled          bool          `yaml:"enabled" mapstructure:"enabled"`
	ReminderInterval time.Duration `yaml:"reminder_interval" mapstructure:"reminder_interval"`
	SoundEnabled     bool          `yaml:"sound_enabled" mapstructure:"sound_enabled"`
}

// StorageConfig holds local storage configuration
type StorageConfig struct {
	DatabasePath   string        `yaml:"database_path" mapstructure:"database_path"`
	BackupEnabled  bool          `yaml:"backup_enabled" mapstructure:"backup_enabled"`
	BackupInterval time.Duration `yaml:"backup_interval" mapstructure:"backup_interval"`
	MaxSessions    int           `yaml:"max_sessions" mapstructure:"max_sessions"`
	CompressDB     bool          `yaml:"compress_db" mapstructure:"compress_db"`
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
		Storage: StorageConfig{
			DatabasePath:   path.Join(pkg.UnixCraftieConfigDir, "sessions.db"),
			BackupEnabled:  true,
			BackupInterval: 24 * time.Hour,
			MaxSessions:    10000,
			CompressDB:     true,
		},
		Logging: LoggingConfig{
			Level:      "info",
			OutputFile: path.Join(DefaultConfigPath(), "craftie.log"),
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

	if c.Storage.DatabasePath == "" {
		return pkg.NewValidationError("storage.database_path is required")
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
