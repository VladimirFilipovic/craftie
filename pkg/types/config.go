package types

import "time"

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
	CredentialsFile string        `yaml:"credentials_file" mapstructure:"credentials_file"`
	SpreadsheetID   string        `yaml:"spreadsheet_id" mapstructure:"spreadsheet_id"`
	SheetName       string        `yaml:"sheet_name" mapstructure:"sheet_name"`
	SyncInterval    time.Duration `yaml:"sync_interval" mapstructure:"sync_interval"`
	Enabled         bool          `yaml:"enabled" mapstructure:"enabled"`
	RetryAttempts   int           `yaml:"retry_attempts" mapstructure:"retry_attempts"`
	RetryDelay      time.Duration `yaml:"retry_delay" mapstructure:"retry_delay"`
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

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		GoogleSheets: GoogleSheetsConfig{
			CredentialsFile: "~/.craftie/service-account.json",
			SpreadsheetID:   "",
			SheetName:       "CraftTime",
			SyncInterval:    15 * time.Minute,
			Enabled:         true,
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
			Level:      "info",
			Format:     "text",
			OutputFile: "~/.craftie/craftie.log",
			MaxSize:    10, // 10MB
			MaxBackups: 3,
			MaxAge:     30, // 30 days
		},
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.GoogleSheets.Enabled {
		if c.GoogleSheets.CredentialsFile == "" {
			return NewValidationError("google_sheets.credentials_file is required when Google Sheets is enabled")
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
