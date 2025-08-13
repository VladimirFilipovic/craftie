# Craftie Implementation Specification

## Go Module Structure and Dependencies

### Go Module Definition (`go.mod`)

```go
module github.com/vlad/craftie

go 1.21

require (
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.2
    github.com/mattn/go-sqlite3 v1.14.19
    google.golang.org/api v0.155.0
    github.com/getlantern/systray v1.2.2
    github.com/gen2brain/beeep v0.0.0-20230907135156-1a38885a97fc
    github.com/sirupsen/logrus v1.9.3
    golang.org/x/oauth2 v0.15.0
)

require (
    // Indirect dependencies will be managed by Go modules
    github.com/fsnotify/fsnotify v1.7.0 // indirect
    github.com/hashicorp/hcl v1.0.0 // indirect
    github.com/inconshreveable/mousetrap v1.1.0 // indirect
    github.com/magiconair/properties v1.8.7 // indirect
    github.com/mitchellh/mapstructure v1.5.0 // indirect
    github.com/pelletier/go-toml/v2 v2.1.1 // indirect
    github.com/spf13/afero v1.11.0 // indirect
    github.com/spf13/cast v1.6.0 // indirect
    github.com/spf13/jwalterweatherman v1.1.0 // indirect
    github.com/spf13/pflag v1.0.5 // indirect
    github.com/subosito/gotenv v1.6.0 // indirect
    golang.org/x/sys v0.15.0 // indirect
    golang.org/x/text v0.14.0 // indirect
    gopkg.in/ini.v1 v1.67.0 // indirect
    gopkg.in/yaml.v3 v3.0.1 // indirect
)
```

### Project Directory Structure

```
craftie/
├── cmd/
│   └── craftie/
│       ├── main.go                 # Application entry point
│       ├── root.go                 # Root command setup
│       ├── start.go                # Start command implementation
│       ├── stop.go                 # Stop command implementation
│       ├── status.go               # Status command implementation
│       └── exit.go                 # Exit command implementation
├── internal/
│   ├── daemon/
│   │   ├── daemon.go               # Main daemon service
│   │   ├── ipc.go                  # Inter-process communication
│   │   ├── server.go               # IPC server implementation
│   │   └── client.go               # IPC client implementation
│   ├── session/
│   │   ├── manager.go              # Session management logic
│   │   ├── timer.go                # Timer and duration tracking
│   │   └── state.go                # Session state management
│   ├── storage/
│   │   ├── sqlite.go               # SQLite database operations
│   │   ├── models.go               # Data models and structs
│   │   ├── migrations.go           # Database schema migrations
│   │   └── queries.go              # SQL queries and operations
│   ├── sheets/
│   │   ├── client.go               # Google Sheets API client
│   │   ├── auth.go                 # Service account authentication
│   │   ├── sync.go                 # Data synchronization logic
│   │   └── formatter.go            # Data formatting for sheets
│   ├── notifications/
│   │   ├── tray.go                 # System tray implementation
│   │   ├── desktop.go              # Desktop notification system
│   │   ├── manager.go              # Notification management
│   │   └── icons.go                # Icon resources and management
│   └── config/
│       ├── config.go               # Configuration management
│       ├── defaults.go             # Default configuration values
│       └── validation.go           # Configuration validation
├── pkg/
│   └── types/
│       ├── session.go              # Session data types
│       ├── config.go               # Configuration data types
│       └── errors.go               # Custom error types
├── scripts/
│   ├── install.sh                  # Installation script
│   ├── uninstall.sh               # Uninstallation script
│   └── service/
│       ├── craftie.service         # systemd service file
│       ├── craftie.plist           # macOS LaunchAgent plist
│       └── craftie.bat             # Windows service batch
├── assets/
│   ├── icons/
│   │   ├── craftie.ico             # Windows icon
│   │   ├── craftie.png             # Linux icon
│   │   └── craftie.icns            # macOS icon
│   └── templates/
│       └── config.yaml.template    # Configuration template
├── docs/
│   ├── setup.md                    # Setup and installation guide
│   ├── usage.md                    # Usage documentation
│   ├── api.md                      # Internal API documentation
│   └── troubleshooting.md          # Troubleshooting guide
├── tests/
│   ├── integration/
│   │   ├── daemon_test.go          # Daemon integration tests
│   │   └── sheets_test.go          # Sheets integration tests
│   └── unit/
│       ├── session_test.go         # Session unit tests
│       ├── storage_test.go         # Storage unit tests
│       └── config_test.go          # Configuration unit tests
├── .gitignore
├── .goreleaser.yml                 # Release configuration
├── Dockerfile                      # Container build file
├── Makefile                        # Build automation
├── README.md                       # Project documentation
├── LICENSE                         # License file
├── go.mod                          # Go module definition
└── go.sum                          # Go module checksums
```

## Core Data Types and Interfaces

### Session Types (`pkg/types/session.go`)

```go
package types

import (
    "time"
)

// Session represents a craft time tracking session
type Session struct {
    ID          int64     `json:"id" db:"id"`
    StartTime   time.Time `json:"start_time" db:"start_time"`
    EndTime     *time.Time `json:"end_time,omitempty" db:"end_time"`
    Duration    int64     `json:"duration" db:"duration"` // seconds
    ProjectName string    `json:"project_name" db:"project_name"`
    Notes       string    `json:"notes" db:"notes"`
    SyncedToSheets bool   `json:"synced_to_sheets" db:"synced_to_sheets"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// SessionStatus represents the current status of the tracking system
type SessionStatus struct {
    IsActive      bool      `json:"is_active"`
    CurrentSession *Session `json:"current_session,omitempty"`
    TotalSessions int64     `json:"total_sessions"`
    TodayDuration int64     `json:"today_duration"` // seconds
}

// SessionManager interface defines session management operations
type SessionManager interface {
    StartSession(projectName, notes string) (*Session, error)
    StopSession() (*Session, error)
    GetCurrentSession() (*Session, error)
    GetStatus() (*SessionStatus, error)
    GetSessions(limit, offset int) ([]*Session, error)
}
```

### Configuration Types (`pkg/types/config.go`)

```go
package types

import "time"

// Config represents the application configuration
type Config struct {
    GoogleSheets GoogleSheetsConfig `yaml:"google_sheets" mapstructure:"google_sheets"`
    Notifications NotificationConfig `yaml:"notifications" mapstructure:"notifications"`
    Storage      StorageConfig      `yaml:"storage" mapstructure:"storage"`
    Daemon       DaemonConfig       `yaml:"daemon" mapstructure:"daemon"`
    Logging      LoggingConfig      `yaml:"logging" mapstructure:"logging"`
}

// GoogleSheetsConfig holds Google Sheets API configuration
type GoogleSheetsConfig struct {
    CredentialsFile string `yaml:"credentials_file" mapstructure:"credentials_file"`
    SpreadsheetID   string `yaml:"spreadsheet_id" mapstructure:"spreadsheet_id"`
    SheetName       string `yaml:"sheet_name" mapstructure:"sheet_name"`
    SyncInterval    time.Duration `yaml:"sync_interval" mapstructure:"sync_interval"`
}

// NotificationConfig holds notification system configuration
type NotificationConfig struct {
    Enabled          bool          `yaml:"enabled" mapstructure:"enabled"`
    ReminderInterval time.Duration `yaml:"reminder_interval" mapstructure:"reminder_interval"`
    ShowTrayIcon     bool          `yaml:"show_tray_icon" mapstructure:"show_tray_icon"`
    SoundEnabled     bool          `yaml:"sound_enabled" mapstructure:"sound_enabled"`
}

// StorageConfig holds local storage configuration
type StorageConfig struct {
    DatabasePath string `yaml:"database_path" mapstructure:"database_path"`
    BackupEnabled bool  `yaml:"backup_enabled" mapstructure:"backup_enabled"`
    BackupInterval time.Duration `yaml:"backup_interval" mapstructure:"backup_interval"`
}

// DaemonConfig holds daemon service configuration
type DaemonConfig struct {
    SocketPath string `yaml:"socket_path" mapstructure:"socket_path"`
    PidFile    string `yaml:"pid_file" mapstructure:"pid_file"`
    LogFile    string `yaml:"log_file" mapstructure:"log_file"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
    Level      string `yaml:"level" mapstructure:"level"`
    Format     string `yaml:"format" mapstructure:"format"`
    OutputFile string `yaml:"output_file" mapstructure:"output_file"`
}
```

## Key Implementation Details

### IPC Communication Protocol

- **Transport**: Unix domain sockets (Linux/macOS), Named pipes (Windows)
- **Protocol**: JSON-based request/response
- **Commands**: START, STOP, STATUS, EXIT
- **Security**: Socket permissions restricted to user only

### Database Schema

```sql
-- Sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    duration INTEGER DEFAULT 0,
    project_name TEXT NOT NULL DEFAULT '',
    notes TEXT DEFAULT '',
    synced_to_sheets BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_sessions_start_time ON sessions(start_time);
CREATE INDEX IF NOT EXISTS idx_sessions_synced ON sessions(synced_to_sheets);
CREATE INDEX IF NOT EXISTS idx_sessions_project ON sessions(project_name);

-- Configuration table for runtime settings
CREATE TABLE IF NOT EXISTS app_config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Google Sheets Integration

- **Authentication**: Service Account with JSON key file
- **API Scope**: `https://www.googleapis.com/auth/spreadsheets`
- **Batch Operations**: Group multiple rows for efficiency
- **Error Handling**: Exponential backoff with jitter
- **Rate Limiting**: Respect API quotas (100 requests/100 seconds/user)

### System Tray Integration

- **Cross-platform**: Different implementations per OS
- **Menu Items**: Status, Start/Stop, Exit
- **Icon States**: Idle, Active, Syncing, Error
- **Notifications**: Native desktop notifications

### Build and Deployment

```makefile
# Makefile targets
.PHONY: build test clean install uninstall

build:
	go build -o bin/craftie cmd/craftie/main.go

test:
	go test ./...

install:
	sudo cp bin/craftie /usr/local/bin/
	mkdir -p ~/.craftie
	cp assets/templates/config.yaml.template ~/.craftie/config.yaml

clean:
	rm -rf bin/
	go clean

release:
	goreleaser release --rm-dist
```

## Security and Performance Considerations

### Security

1. **File Permissions**: Config and credential files set to 600 (user read/write only)
2. **IPC Security**: Socket/pipe permissions restricted to user
3. **Credential Storage**: Service account JSON stored securely
4. **Input Validation**: All user inputs validated and sanitized

### Performance

1. **Memory Usage**: Target <50MB RAM usage for daemon
2. **Database**: SQLite with WAL mode for concurrent access
3. **Network**: Batch API calls, connection pooling
4. **Startup Time**: Target <2 seconds for daemon startup
5. **Resource Cleanup**: Proper goroutine and connection management

### Error Recovery

1. **Database Corruption**: Automatic backup and recovery
2. **Network Failures**: Offline mode with sync queue
3. **API Rate Limits**: Exponential backoff and retry
4. **Daemon Crashes**: Auto-restart via system service
5. **Configuration Errors**: Fallback to defaults with warnings
