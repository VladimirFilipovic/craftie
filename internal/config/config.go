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

func (cm *ConfigManager) LoadConfig(configPath string) (*Config, error) {
	var config *Config

	generateDefaultConfig := false
	if configPath == "" {
		configPath = DefaultConfigPath()
		generateDefaultConfig = true
	}

	fullConfigPath, err := pkg.GetExpandedPathWithHome(configPath)
	if err != nil {
		return nil, err
	}

	// load existing config file
	if !generateDefaultConfig {
		// Load existing config file
		cm.viper.AddConfigPath(filepath.Dir(fullConfigPath))
		cm.viper.SetConfigName(filepath.Base(fullConfigPath))
		cm.viper.SetConfigType("yaml")
		cm.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		cm.viper.AutomaticEnv()
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		config = defaultConfig()
		if err := cm.viper.Unmarshal(config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

	}

	// When no config path is provided or default configPath is provided
	// check if default config file exists and if not create it with sensible values
	if _, err := os.Stat(fullConfigPath); generateDefaultConfig && os.IsNotExist(err) {
		// Create default config file
		config = defaultConfig()
		if err := createConfigFile(fullConfigPath, config); err != nil {
			return nil, fmt.Errorf("failed to create default config file: %w", err)
		}
		fmt.Printf("Default config file created at %s\n", fullConfigPath)
	}

	expandPaths(config)

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
	ShowTrayIcon     bool          `yaml:"show_tray_icon" mapstructure:"show_tray_icon"`
	SoundEnabled     bool          `yaml:"sound_enabled" mapstructure:"sound_enabled"`
	ReminderMessage  string        `yaml:"reminder_message" mapstructure:"reminder_message"`
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
	Format     string `yaml:"format" mapstructure:"format"`
	OutputFile string `yaml:"output_file" mapstructure:"output_file"`
	// MaxSize    int    `yaml:"max_size" mapstructure:"max_size"`       // MB
	// MaxBackups int    `yaml:"max_backups" mapstructure:"max_backups"` // number of backup files
	// MaxAge     int    `yaml:"max_age" mapstructure:"max_age"`         // days
}

func defaultConfig() *Config {
	return &Config{
		GoogleSheets: GoogleSheetsConfig{
			Enabled: false,
		},
		Notifications: NotificationConfig{
			Enabled:          true,
			ReminderInterval: 1 * time.Hour,
			ShowTrayIcon:     true,
			SoundEnabled:     false,
			ReminderMessage:  "Craftie is tracking your time - %s elapsed",
		},
		Storage: StorageConfig{
			DatabasePath:   path.Join(pkg.UnixCraftieConfigDir, "sessions.db"),
			BackupEnabled:  true,
			BackupInterval: 24 * time.Hour,
			MaxSessions:    10000,
			CompressDB:     true,
		},
		DaemonSocketPath: "~/.craftie/daemon.sock",
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			OutputFile: path.Join(DefaultConfigPath(), "craftie.log"),
			// MaxSize:    10, // 10MB
			// MaxBackups: 3,
			// MaxAge:     30, // 30 days
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
	// Validate log format
	validFormats := map[string]bool{"text": true, "json": true}
	if !validFormats[c.Logging.Format] {
		return pkg.NewValidationError("logging.format must be either 'text' or 'json'")
	}

	return nil
}
