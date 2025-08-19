package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"github.com/vlad/craftie/internal/path"
	"github.com/vlad/craftie/pkg/types"
)

type ConfigManager struct {
	config *types.Config
	viper  *viper.Viper
}

func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		viper: viper.New(),
	}
}

/*
	 Loads configuration from file and environment variables using viper.
	 Viper uses the following precedence order.
	 Each item takes precedence over the item below it:
		explicit call to Set
		flag
		env
		config
		key/value store
	  default
*/
func (m *ConfigManager) Load(configPath string) error {
	// Set default configuration
	m.config = types.DefaultConfig()

	configPath, err := path.ExpandPathWithHome(configPath)

	if err != nil {
		return &types.CraftieError{
			Code:    types.ErrCodeConfig,
			Message: "Path failed to expand",
			Cause:   err,
		}
	}

	// Set up config file
	m.viper.SetConfigFile(configPath)
	m.viper.SetConfigType("yaml")

	// Set environment variable prefix
	m.viper.SetEnvPrefix("CRAFTIE")
	m.viper.AutomaticEnv()
	m.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Try to read config file
	if err := m.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, create default one
			if err := m.createDefaultConfig(configPath); err != nil {
				return err
			}
		}
		return types.NewValidationError("failed to read config file: " + err.Error())
	}

	// Unmarshal into config struct
	if err := m.viper.Unmarshal(m.config); err != nil {
		return types.NewValidationError("failed to parse config: " + err.Error())
	}

	// Expand paths in config
	m.expandPaths()

	// Validate configuration
	if err := m.config.Validate(); err != nil {
		return err
	}

	return nil
}

func (m *ConfigManager) GetConfig() *types.Config {
	return m.config
}

func (m *ConfigManager) Save(configPath string) error {
	configPath, err := path.ExpandPathWithHome(configPath)

	if err != nil {
		return &types.CraftieError{
			Code:    types.ErrCodeConfig,
			Message: "Path issues",
			Cause:   err,
		}
	}

	m.viper.SetConfigFile(configPath)
	return m.viper.WriteConfig()
}

// Set sets a configuration value
func (m *ConfigManager) Set(key string, value interface{}) {
	m.viper.Set(key, value)
	// Re-unmarshal to update the config struct
	m.viper.Unmarshal(m.config)
}

// Get gets a configuration value
func (m *ConfigManager) Get(key string) interface{} {
	return m.viper.Get(key)
}

// createDefaultConfig creates a default configuration file
func (m *ConfigManager) createDefaultConfig(configPath string) error {
	// Set default values in viper
	m.setDefaults()

	// Write config file
	if err := m.viper.SafeWriteConfigAs(configPath); err != nil {
		return types.NewValidationError("failed to create default config: " + err.Error())
	}

	return nil
}

// setDefaults sets default values in viper
func (m *ConfigManager) setDefaults() {
	defaults := types.DefaultConfig()

	// Google Sheets defaults
	m.viper.SetDefault("google_sheets.credentials_file", defaults.GoogleSheets.CredentialsFile)
	m.viper.SetDefault("google_sheets.spreadsheet_id", defaults.GoogleSheets.SpreadsheetID)
	m.viper.SetDefault("google_sheets.sheet_name", defaults.GoogleSheets.SheetName)
	m.viper.SetDefault("google_sheets.sync_interval", defaults.GoogleSheets.SyncInterval)
	m.viper.SetDefault("google_sheets.enabled", defaults.GoogleSheets.Enabled)
	m.viper.SetDefault("google_sheets.retry_attempts", defaults.GoogleSheets.RetryAttempts)
	m.viper.SetDefault("google_sheets.retry_delay", defaults.GoogleSheets.RetryDelay)

	// Notifications defaults
	m.viper.SetDefault("notifications.enabled", defaults.Notifications.Enabled)
	m.viper.SetDefault("notifications.reminder_interval", defaults.Notifications.ReminderInterval)
	m.viper.SetDefault("notifications.show_tray_icon", defaults.Notifications.ShowTrayIcon)
	m.viper.SetDefault("notifications.sound_enabled", defaults.Notifications.SoundEnabled)
	m.viper.SetDefault("notifications.reminder_message", defaults.Notifications.ReminderMessage)

	// Storage defaults
	m.viper.SetDefault("storage.database_path", defaults.Storage.DatabasePath)
	m.viper.SetDefault("storage.backup_enabled", defaults.Storage.BackupEnabled)
	m.viper.SetDefault("storage.backup_interval", defaults.Storage.BackupInterval)
	m.viper.SetDefault("storage.max_sessions", defaults.Storage.MaxSessions)
	m.viper.SetDefault("storage.compress_db", defaults.Storage.CompressDB)

	// Daemon defaults
	m.viper.SetDefault("daemon.socket_path", defaults.Daemon.SocketPath)
	m.viper.SetDefault("daemon.pid_file", defaults.Daemon.PidFile)
	m.viper.SetDefault("daemon.log_file", defaults.Daemon.LogFile)
	m.viper.SetDefault("daemon.auto_start", defaults.Daemon.AutoStart)

	// Logging defaults
	m.viper.SetDefault("logging.level", defaults.Logging.Level)
	m.viper.SetDefault("logging.format", defaults.Logging.Format)
	m.viper.SetDefault("logging.output_file", defaults.Logging.OutputFile)
	m.viper.SetDefault("logging.max_size", defaults.Logging.MaxSize)
	m.viper.SetDefault("logging.max_backups", defaults.Logging.MaxBackups)
	m.viper.SetDefault("logging.max_age", defaults.Logging.MaxAge)
}

// expandPaths expands ~ to home directory in file paths
func (m *ConfigManager) expandPaths() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return // Skip expansion if we can't get home dir
	}

	// Expand Google Sheets credentials file path
	if strings.HasPrefix(m.config.GoogleSheets.CredentialsFile, "~/") {
		m.config.GoogleSheets.CredentialsFile = filepath.Join(homeDir, m.config.GoogleSheets.CredentialsFile[2:])
	}

	// Expand storage database path
	if strings.HasPrefix(m.config.Storage.DatabasePath, "~/") {
		m.config.Storage.DatabasePath = filepath.Join(homeDir, m.config.Storage.DatabasePath[2:])
	}

	// Expand daemon paths
	if strings.HasPrefix(m.config.Daemon.SocketPath, "~/") {
		m.config.Daemon.SocketPath = filepath.Join(homeDir, m.config.Daemon.SocketPath[2:])
	}
	if strings.HasPrefix(m.config.Daemon.PidFile, "~/") {
		m.config.Daemon.PidFile = filepath.Join(homeDir, m.config.Daemon.PidFile[2:])
	}
	if strings.HasPrefix(m.config.Daemon.LogFile, "~/") {
		m.config.Daemon.LogFile = filepath.Join(homeDir, m.config.Daemon.LogFile[2:])
	}

	// Expand logging output file path
	if strings.HasPrefix(m.config.Logging.OutputFile, "~/") {
		m.config.Logging.OutputFile = filepath.Join(homeDir, m.config.Logging.OutputFile[2:])
	}
}

// ValidateGoogleSheetsConfig validates Google Sheets configuration
func (m *ConfigManager) ValidateGoogleSheetsConfig() error {
	if !m.config.GoogleSheets.Enabled {
		return nil // Skip validation if disabled
	}

	// Check if credentials file exists
	if _, err := os.Stat(m.config.GoogleSheets.CredentialsFile); os.IsNotExist(err) {
		return types.NewValidationError("Google Sheets credentials file not found: " + m.config.GoogleSheets.CredentialsFile)
	}

	// Check if spreadsheet ID is set
	if m.config.GoogleSheets.SpreadsheetID == "" {
		return types.NewValidationError("Google Sheets spreadsheet ID is required")
	}

	return nil
}

// GetDefaultConfigPath returns the default configuration file path
func GetDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".craftie/config.yaml" // Fallback to relative path
	}
	return filepath.Join(homeDir, ".craftie", "config.yaml")
}

// CreateConfigTemplate creates a configuration template file
func CreateConfigTemplate(path string) error {
	template := `# Craftie Configuration File
# This file contains all configuration options for Craftie

google_sheets:
  # Path to Google Service Account credentials JSON file
  credentials_file: "~/.craftie/service-account.json"
  
  # Google Sheets spreadsheet ID (found in the URL)
  spreadsheet_id: ""
  
  # Name of the sheet/tab to write to
  sheet_name: "CraftTime"
  
  # How often to sync data to Google Sheets
  sync_interval: "15m"
  
  # Enable/disable Google Sheets integration
  enabled: true
  
  # Number of retry attempts for failed API calls
  retry_attempts: 3
  
  # Delay between retry attempts
  retry_delay: "5s"

notifications:
  # Enable/disable all notifications
  enabled: true
  
  # How often to show reminder notifications
  reminder_interval: "1h"
  
  # Show system tray icon
  show_tray_icon: true
  
  # Enable notification sounds
  sound_enabled: false
  
  # Message template for reminders (%s will be replaced with elapsed time)
  reminder_message: "Craftie is tracking your time - %s elapsed"

storage:
  # Path to SQLite database file
  database_path: "~/.craftie/sessions.db"
  
  # Enable automatic database backups
  backup_enabled: true
  
  # How often to create backups
  backup_interval: "24h"
  
  # Maximum number of sessions to keep (0 = unlimited)
  max_sessions: 10000
  
  # Enable database compression
  compress_db: false

daemon:
  # Path to Unix socket for IPC communication
  socket_path: "~/.craftie/daemon.sock"
  
  # Path to PID file
  pid_file: "~/.craftie/craftie.pid"
  
  # Path to daemon log file
  log_file: "~/.craftie/craftie.log"
  
  # Auto-start daemon on system boot
  auto_start: false

logging:
  # Log level: trace, debug, info, warn, error, fatal, panic
  level: "info"
  
  # Log format: text or json
  format: "text"
  
  # Path to log file

  output_file: "~/.craftie/craftie.log"
  
  # Maximum log file size in MB
  max_size: 10
  
  # Maximum number of backup log files
  max_backups: 3
  
  # Maximum age of log files in days
  max_age: 30
`

	// Expand home directory if needed
	if path[:2] == "~/" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return types.NewValidationError("failed to get home directory")
		}
		path = filepath.Join(homeDir, path[2:])
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return types.NewValidationError("failed to create config directory: " + err.Error())
	}

	// Write template file
	if err := os.WriteFile(path, []byte(template), 0644); err != nil {
		return types.NewValidationError("failed to write config template: " + err.Error())
	}

	return nil
}
