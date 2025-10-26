package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/vlad/craftie/internal/path"
)

type ConfigManager struct {
	Config *Config
	viper  *viper.Viper
}

func NewConfigManager(configPath string) (*ConfigManager, error) {
	viper := viper.New()

	config, err := loadConfig(configPath, viper)

	if err != nil {
		return nil, fmt.Errorf("failed to initialize config manager: %w", err)
	}

	return &ConfigManager{
		viper:  viper,
		Config: config,
	}, nil
}

func loadConfig(configPath string, viper *viper.Viper) (*Config, error) {
	var config *Config

	if configPath == "" {
		config = defaultConfig()
	}

	if configPath != "" {
		fullConfigPath, err := path.ExpandPathWithHome(configPath)

		if err != nil {
			return nil, err
		}

		viper.AddConfigPath(fullConfigPath)
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err != nil {
			return nil, err
		}

		if err := viper.Unmarshal(config); err != nil {
			return nil, err
		}
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
	// Expand Google Sheets credentials file path
	expandedPath, err := path.ExpandPathWithHome(conf.GoogleSheets.CredentialsFile)
	if err != nil {
		return err
	}
	conf.GoogleSheets.CredentialsFile = expandedPath

	// Expand storage database path
	expandedPath, err = path.ExpandPathWithHome(conf.Storage.DatabasePath)
	if err != nil {
		return err
	}
	conf.Storage.DatabasePath = expandedPath

	// Expand daemon paths
	expandedPath, err = path.ExpandPathWithHome(conf.Daemon.SocketPath)
	if err != nil {
		return err
	}
	conf.Daemon.SocketPath = expandedPath
	expandedPath, err = path.ExpandPathWithHome(conf.Daemon.PidFile)
	if err != nil {
		return err
	}
	conf.Daemon.PidFile = expandedPath
	expandedPath, err = path.ExpandPathWithHome(conf.Daemon.LogFile)
	if err != nil {
		return err
	}
	conf.Daemon.LogFile = expandedPath

	// Expand logging output file path
	expandedPath, err = path.ExpandPathWithHome(conf.Logging.OutputFile)
	if err != nil {
		return err
	}
	conf.Logging.OutputFile = expandedPath

	return nil
}

func defaultConfigPath() string {
	return "~/"
}

// CraftieError represents a base error type for the application
type CraftieError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Cause   error  `json:"cause,omitempty"`
}

func (e *CraftieError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *CraftieError) Unwrap() error {
	return e.Cause
}

// Error codes
const (
	ErrCodeValidation    = "VALIDATION_ERROR"
	ErrCodeDatabase      = "DATABASE_ERROR"
	ErrCodeNetwork       = "NETWORK_ERROR"
	ErrCodeAuth          = "AUTH_ERROR"
	ErrCodeConfig        = "CONFIG_ERROR"
	ErrCodeSession       = "SESSION_ERROR"
	ErrCodeDaemon        = "DAEMON_ERROR"
	ErrCodeNotification  = "NOTIFICATION_ERROR"
	ErrCodeSync          = "SYNC_ERROR"
	ErrCodeFileSystem    = "FILESYSTEM_ERROR"
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeAlreadyExists = "ALREADY_EXISTS"
	ErrCodePermission    = "PERMISSION_ERROR"
	ErrCodeTimeout       = "TIMEOUT_ERROR"
)

// ValidationError represents a configuration or input validation error
type ValidationError struct {
	*CraftieError
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		CraftieError: &CraftieError{
			Code:    ErrCodeValidation,
			Message: message,
		},
	}
}

// Config represents the application configuration
type Config struct {
	GoogleSheets  GoogleSheetsConfig `yaml:"google_sheets" mapstructure:"google_sheets"`
	Notifications NotificationConfig `yaml:"notifications" mapstructure:"notifications"`
	Storage       StorageConfig      `yaml:"storage" mapstructure:"storage"`
	Daemon        DaemonConfig       `yaml:"daemon" mapstructure:"daemon"`
	Logging       LoggingConfig      `yaml:"logging" mapstructure:"logging"`
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

// DaemonConfig holds daemon service configuration
type DaemonConfig struct {
	SocketPath string `yaml:"socket_path" mapstructure:"socket_path"`
	PidFile    string `yaml:"pid_file" mapstructure:"pid_file"`
	LogFile    string `yaml:"log_file" mapstructure:"log_file"`
	AutoStart  bool   `yaml:"auto_start" mapstructure:"auto_start"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level" mapstructure:"level"`
	Format     string `yaml:"format" mapstructure:"format"`
	OutputFile string `yaml:"output_file" mapstructure:"output_file"`
	MaxSize    int    `yaml:"max_size" mapstructure:"max_size"`       // MB
	MaxBackups int    `yaml:"max_backups" mapstructure:"max_backups"` // number of backup files
	MaxAge     int    `yaml:"max_age" mapstructure:"max_age"`         // days
}

func defaultConfig() *Config {
	return &Config{
		GoogleSheets: GoogleSheetsConfig{
			CredentialsFile: "~/.craftie/service-account.json",
			SpreadsheetID:   "",
			SheetName:       "CraftTime",
			SyncInterval:    15 * time.Minute,
			Enabled:         false,
			RetryAttempts:   3,
			RetryDelay:      5 * time.Second,
		},
		Notifications: NotificationConfig{
			Enabled:          true,
			ReminderInterval: 1 * time.Hour,
			ShowTrayIcon:     true,
			SoundEnabled:     false,
			ReminderMessage:  "Craftie is tracking your time - %s elapsed",
		},
		Storage: StorageConfig{
			DatabasePath:   "~/.craftie/sessions.db",
			BackupEnabled:  true,
			BackupInterval: 24 * time.Hour,
			MaxSessions:    10000,
			CompressDB:     false,
		},
		Daemon: DaemonConfig{
			SocketPath: "~/.craftie/daemon.sock",
			PidFile:    "~/.craftie/craftie.pid",
			LogFile:    "~/.craftie/craftie.log",
			AutoStart:  false,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
			// TODO: Wont log into file. Gonna be streaming to stdout instead. And call it a day
			OutputFile: "~/.craftie/craftie.log",
			MaxSize:    10, // 10MB
			MaxBackups: 3,
			MaxAge:     30, // 30 days
		},
	}
}

func (c *Config) Validate() error {
	if c.GoogleSheets.Enabled {
		if c.GoogleSheets.CredentialsFile == "" {
			return NewValidationError("google_sheets.credentials_file is required when Google Sheets is enabled")
		}

		// Check if credentials file exists
		if _, err := os.Stat(c.GoogleSheets.CredentialsFile); os.IsNotExist(err) {
			return NewValidationError("Google Sheets credentials file not found: " + c.GoogleSheets.CredentialsFile)
		}

		if c.GoogleSheets.SpreadsheetID == "" {
			return NewValidationError("google_sheets.spreadsheet_id is required when Google Sheets is enabled")
		}
	}

	if c.Storage.DatabasePath == "" {
		return NewValidationError("storage.database_path is required")
	}

	if c.Daemon.SocketPath == "" {
		return NewValidationError("daemon.socket_path is required")
	}

	if c.Daemon.PidFile == "" {
		return NewValidationError("daemon.pid_file is required")
	}

	// Validate log level
	validLevels := map[string]bool{
		"trace": true, "debug": true, "info": true,
		"warn": true, "error": true, "fatal": true, "panic": true,
	}
	if !validLevels[c.Logging.Level] {
		return NewValidationError("logging.level must be one of: trace, debug, info, warn, error, fatal, panic")
	}

	// Validate log format
	validFormats := map[string]bool{"text": true, "json": true}
	if !validFormats[c.Logging.Format] {
		return NewValidationError("logging.format must be either 'text' or 'json'")
	}

	return nil
}
