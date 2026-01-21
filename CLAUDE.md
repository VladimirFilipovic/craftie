# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run Commands

```bash
# Build the binary
go build -o craftie ./cmd/craftie.go

# Run tests
go test ./...

# Run a single test
go test -run TestExecuteCredentialsHelper ./internal/sheets/

# Run the application
./craftie start -p "project-name" -e 2h
```

## Architecture Overview

Craftie is a CLI time tracking application written in Go that records work sessions and optionally syncs them to Google Sheets or exports to CSV.

### Core Flow

1. **Entry point** (`cmd/craftie.go`): Uses urfave/cli/v3 for command handling. The `start` command creates a session and enters a main loop that handles signals (SIGINT/SIGTERM), optional end timers, and periodic syncing.

2. **Session management** (`internal/session/`): The `Session` struct tracks start/end times, project name, and notes. Sessions can have optional duration limits set via `-e` flag (e.g., `2h`, `30m`).

3. **Data persistence** (`internal/sheets/`):
   - `google.go`: Creates/updates rows in Google Sheets using service account authentication
   - `csv.go`: Writes to local CSV files with in-place row updates via file truncation
   - `credentials.go`: Fetches Google credentials from a helper script or system keyring (go-keyring)

4. **Sync pattern**: Sessions are saved immediately on start, then periodically synced (per `config.SessionSyncTime`), and finally synced on session end. Both Google Sheets and CSV use a "sync state" pattern that tracks the row location for in-place updates.

### Configuration

Config is loaded from `~/.config/craftie/craftie.yaml` (XDG_CONFIG_HOME respected). See `config-template.yaml` for all options. Key integrations:

- `google_sheets`: Requires spreadsheet_id, sheet_name, and credentials via helper script or keyring
- `csv`: Requires file_path when enabled
- Paths support `~` expansion

### Credentials

Google Sheets authentication uses service account JWT. Credentials can come from:

1. A credentials helper script (configured via `credentials_helper` path) that outputs the JSON
2. System keyring under service "craftie", key "google-sheets"

## Style guide

Write concise simple code.
Avoid to many if else statements i prefer not to have nesting for conditionals.
Write comments only for crucial parts of the code that need further clarification.
Comments should usually explain _WHY_ not _WHAT_.
