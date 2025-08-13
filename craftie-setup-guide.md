# Craftie Setup and Usage Guide

## Prerequisites

### System Requirements

- **Operating System**: Linux, macOS, or Windows
- **Go Version**: 1.21 or higher
- **Memory**: Minimum 100MB RAM
- **Storage**: 50MB for application and database
- **Network**: Internet connection for Google Sheets sync

### Google Sheets Setup

1. **Create Google Cloud Project**:

   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project or select existing one
   - Enable Google Sheets API

2. **Create Service Account**:

   - Navigate to IAM & Admin > Service Accounts
   - Click "Create Service Account"
   - Name: `craftie-tracker`
   - Description: `Service account for Craftie time tracking`

3. **Generate Credentials**:

   - Click on created service account
   - Go to "Keys" tab
   - Click "Add Key" > "Create new key"
   - Select JSON format
   - Download and save as `service-account.json`

4. **Prepare Google Sheet**:
   - Create new Google Sheet
   - Name it "Craft Time Tracking" (or your preference)
   - Add headers in row 1:
     - A1: "Timestamp"
     - B1: "Duration"
     - C1: "Project/Task Name"
     - D1: "Notes"
   - Share sheet with service account email (found in JSON file)
   - Give "Editor" permissions
   - Copy the Sheet ID from URL

## Installation Process

### Option 1: Build from Source

```bash
# Clone repository
git clone https://github.com/vlad/craftie.git
cd craftie

# Build application
make build

# Install system-wide
sudo make install

# Or install for current user only
make install-user
```

### Option 2: Download Pre-built Binary

```bash
# Download latest release
curl -L https://github.com/vlad/craftie/releases/latest/download/craftie-linux-amd64.tar.gz | tar xz

# Move to PATH
sudo mv craftie /usr/local/bin/

# Make executable
sudo chmod +x /usr/local/bin/craftie
```

### Option 3: Using Go Install

```bash
go install github.com/vlad/craftie/cmd/craftie@latest
```

## Initial Configuration

### 1. Create Configuration Directory

```bash
mkdir -p ~/.craftie
```

### 2. Setup Configuration File

Create `~/.craftie/config.yaml`:

```yaml
google_sheets:
  credentials_file: "~/.craftie/service-account.json"
  spreadsheet_id: "YOUR_SHEET_ID_HERE"
  sheet_name: "Sheet1"
  sync_interval: "15m"

notifications:
  enabled: true
  reminder_interval: "1h"
  show_tray_icon: true
  sound_enabled: false

storage:
  database_path: "~/.craftie/sessions.db"
  backup_enabled: true
  backup_interval: "24h"

daemon:
  socket_path: "~/.craftie/daemon.sock"
  pid_file: "~/.craftie/craftie.pid"
  log_file: "~/.craftie/craftie.log"

logging:
  level: "info"
  format: "text"
  output_file: "~/.craftie/craftie.log"
```

### 3. Place Service Account Credentials

```bash
# Copy your downloaded service account JSON
cp /path/to/your/service-account.json ~/.craftie/service-account.json

# Secure the credentials file
chmod 600 ~/.craftie/service-account.json
```

### 4. Initialize Database

```bash
# Start daemon (will create database automatically)
craftie daemon start

# Verify daemon is running
craftie status
```

## Usage Instructions

### Basic Commands

#### Start Tracking Session

```bash
# Start with project name
craftie start "Web Development"

# Start with project name and notes
craftie start "Web Development" "Working on user authentication"

# Start with just notes (project will be "General")
craftie start "" "Learning new Go patterns"
```

#### Check Current Status

```bash
# Show current session status
craftie status

# Show detailed status with today's summary
craftie status --detailed
```

#### Stop Current Session

```bash
# Stop current session
craftie stop

# Stop with additional notes
craftie stop --notes "Completed login functionality"
```

#### Daemon Management

```bash
# Start daemon
craftie daemon start

# Stop daemon
craftie daemon stop

# Restart daemon
craftie daemon restart

# Check daemon status
craftie daemon status
```

#### View Session History

```bash
# Show recent sessions
craftie history

# Show last 10 sessions
craftie history --limit 10

# Show sessions for specific date
craftie history --date 2024-01-15

# Show sessions for date range
craftie history --from 2024-01-01 --to 2024-01-31
```

### Advanced Usage

#### Configuration Management

```bash
# Show current configuration
craftie config show

# Validate configuration
craftie config validate

# Reset to defaults
craftie config reset

# Edit configuration
craftie config edit
```

#### Data Management

```bash
# Force sync to Google Sheets
craftie sync

# Export sessions to CSV
craftie export --format csv --output sessions.csv

# Import sessions from CSV
craftie import --file sessions.csv

# Backup database
craftie backup --output backup.db
```

#### System Integration

```bash
# Install as system service (Linux/macOS)
sudo craftie service install

# Enable auto-start
sudo craftie service enable

# Start service
sudo craftie service start

# Check service status
sudo craftie service status
```

## Workflow Examples

### Daily Workflow

```bash
# Morning: Start tracking
craftie start "Project Alpha" "Daily standup and planning"

# Check status anytime
craftie status

# Lunch break: Stop tracking
craftie stop

# Afternoon: Resume or start new session
craftie start "Project Beta" "Code review and bug fixes"

# End of day: Stop tracking
craftie stop --notes "Completed feature implementation"

# Review day's work
craftie history --date today
```

### Project-based Tracking

```bash
# Start work on specific project
craftie start "Mobile App" "Implementing push notifications"

# Switch to different project (automatically stops current)
craftie start "Website Redesign" "Working on responsive layout"

# View project summary
craftie history --project "Mobile App" --from 2024-01-01
```

## Troubleshooting

### Common Issues

#### Daemon Won't Start

```bash
# Check if already running
ps aux | grep craftie

# Check log file
tail -f ~/.craftie/craftie.log

# Remove stale PID file
rm ~/.craftie/craftie.pid

# Try starting with verbose logging
craftie daemon start --verbose
```

#### Google Sheets Sync Failing

```bash
# Check credentials file exists
ls -la ~/.craftie/service-account.json

# Validate credentials
craftie config validate

# Test connection
craftie sync --dry-run

# Check API quotas in Google Cloud Console
```

#### Notifications Not Working

```bash
# Check notification settings
craftie config show | grep notifications

# Test notification system
craftie notify test

# Check system notification permissions
```

#### Database Issues

```bash
# Check database file
ls -la ~/.craftie/sessions.db

# Backup current database
cp ~/.craftie/sessions.db ~/.craftie/sessions.db.backup

# Reset database (WARNING: loses data)
rm ~/.craftie/sessions.db
craftie daemon restart
```

### Log Analysis

```bash
# View recent logs
tail -f ~/.craftie/craftie.log

# Search for errors
grep ERROR ~/.craftie/craftie.log

# View logs with timestamps
craftie logs --follow

# Export logs for support
craftie logs --export support-logs.txt
```

## System Service Setup

### Linux (systemd)

```bash
# Install service
sudo craftie service install

# Service file location: /etc/systemd/system/craftie.service
# Enable auto-start
sudo systemctl enable craftie

# Start service
sudo systemctl start craftie

# Check status
sudo systemctl status craftie
```

### macOS (LaunchAgent)

```bash
# Install launch agent
craftie service install --user

# Service file location: ~/Library/LaunchAgents/com.craftie.daemon.plist
# Load service
launchctl load ~/Library/LaunchAgents/com.craftie.daemon.plist

# Check status
launchctl list | grep craftie
```

### Windows (Service)

```cmd
# Install as Windows service (run as Administrator)
craftie service install

# Start service
sc start craftie

# Check status
sc query craftie
```

## Data Privacy and Security

### Local Data Storage

- All session data stored locally in SQLite database
- Database location: `~/.craftie/sessions.db`
- Automatic backups created daily
- No sensitive data transmitted except to configured Google Sheets

### Google Sheets Integration

- Uses service account authentication (no personal Google account required)
- Only sends session data (timestamp, duration, project, notes)
- Data transmitted over HTTPS
- Service account has minimal required permissions

### Credentials Security

- Service account JSON file stored with restricted permissions (600)
- No passwords or personal information stored
- Configuration file readable only by user
- IPC socket secured to user access only

## Performance Optimization

### Resource Usage

- Typical RAM usage: 20-40MB
- Database size: ~1MB per 1000 sessions
- Network usage: Minimal (only during sync)
- CPU usage: <1% during normal operation

### Optimization Tips

```bash
# Reduce sync frequency for better performance
craftie config set google_sheets.sync_interval 30m

# Disable notifications if not needed
craftie config set notifications.enabled false

# Enable database compression
craftie config set storage.compress true

# Limit history retention
craftie config set storage.max_sessions 10000
```

## Backup and Recovery

### Manual Backup

```bash
# Backup database
cp ~/.craftie/sessions.db ~/backups/craftie-$(date +%Y%m%d).db

# Backup configuration
cp ~/.craftie/config.yaml ~/backups/craftie-config-$(date +%Y%m%d).yaml
```

### Automated Backup

```bash
# Enable automatic backups
craftie config set storage.backup_enabled true
craftie config set storage.backup_interval 24h

# Backup location: ~/.craftie/backups/
```

### Recovery

```bash
# Restore from backup
cp ~/backups/craftie-20240115.db ~/.craftie/sessions.db

# Restart daemon
craftie daemon restart

# Verify data
craftie history --limit 5
```
