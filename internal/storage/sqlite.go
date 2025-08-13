package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/vlad/craftie/pkg/types"
)

// SQLiteStorage implements session storage using SQLite
type SQLiteStorage struct {
	db   *sql.DB
	path string
}

// NewSQLiteStorage creates a new SQLite storage instance
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	// Expand home directory if needed
	if dbPath[:2] == "~/" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, types.NewDatabaseErrorWithCause("failed to get home directory", err)
		}
		dbPath = filepath.Join(homeDir, dbPath[2:])
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, types.NewDatabaseErrorWithCause("failed to create database directory", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, types.NewDatabaseErrorWithCause("failed to open database", err)
	}

	storage := &SQLiteStorage{
		db:   db,
		path: dbPath,
	}

	// Initialize database schema
	if err := storage.migrate(); err != nil {
		db.Close()
		return nil, err
	}

	return storage, nil
}

// migrate creates or updates the database schema
func (s *SQLiteStorage) migrate() error {
	// Create sessions table
	createSessionsTable := `
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
	);`

	if _, err := s.db.Exec(createSessionsTable); err != nil {
		return types.NewDatabaseErrorWithCause("failed to create sessions table", err)
	}

	// Create indexes for performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_sessions_start_time ON sessions(start_time);",
		"CREATE INDEX IF NOT EXISTS idx_sessions_synced ON sessions(synced_to_sheets);",
		"CREATE INDEX IF NOT EXISTS idx_sessions_project ON sessions(project_name);",
		"CREATE INDEX IF NOT EXISTS idx_sessions_active ON sessions(end_time) WHERE end_time IS NULL;",
	}

	for _, indexSQL := range indexes {
		if _, err := s.db.Exec(indexSQL); err != nil {
			return types.NewDatabaseErrorWithCause("failed to create index", err)
		}
	}

	// Create app_config table for runtime settings
	createConfigTable := `
	CREATE TABLE IF NOT EXISTS app_config (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := s.db.Exec(createConfigTable); err != nil {
		return types.NewDatabaseErrorWithCause("failed to create config table", err)
	}

	return nil
}

// CreateSession creates a new session in the database
func (s *SQLiteStorage) CreateSession(session *types.Session) error {
	query := `
	INSERT INTO sessions (start_time, end_time, duration, project_name, notes, synced_to_sheets, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	var endTime interface{}
	if session.EndTime != nil {
		endTime = *session.EndTime
	}

	result, err := s.db.Exec(query,
		session.StartTime,
		endTime,
		session.Duration,
		session.ProjectName,
		session.Notes,
		session.SyncedToSheets,
		session.CreatedAt,
		session.UpdatedAt,
	)

	if err != nil {
		return types.NewDatabaseErrorWithCause("failed to create session", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return types.NewDatabaseErrorWithCause("failed to get session ID", err)
	}

	session.ID = id
	return nil
}

// UpdateSession updates an existing session
func (s *SQLiteStorage) UpdateSession(session *types.Session) error {
	query := `
	UPDATE sessions 
	SET start_time = ?, end_time = ?, duration = ?, project_name = ?, notes = ?, 
	    synced_to_sheets = ?, updated_at = ?
	WHERE id = ?`

	var endTime interface{}
	if session.EndTime != nil {
		endTime = *session.EndTime
	}

	session.UpdatedAt = time.Now()

	result, err := s.db.Exec(query,
		session.StartTime,
		endTime,
		session.Duration,
		session.ProjectName,
		session.Notes,
		session.SyncedToSheets,
		session.UpdatedAt,
		session.ID,
	)

	if err != nil {
		return types.NewDatabaseErrorWithCause("failed to update session", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return types.NewDatabaseErrorWithCause("failed to check update result", err)
	}

	if rowsAffected == 0 {
		return types.NewNotFoundError("session", fmt.Sprintf("%d", session.ID))
	}

	return nil
}

// GetSessionByID retrieves a session by its ID
func (s *SQLiteStorage) GetSessionByID(id int64) (*types.Session, error) {
	query := `
	SELECT id, start_time, end_time, duration, project_name, notes, synced_to_sheets, created_at, updated_at
	FROM sessions WHERE id = ?`

	row := s.db.QueryRow(query, id)
	return s.scanSession(row)
}

// GetActiveSession returns the currently active session (no end time)
func (s *SQLiteStorage) GetActiveSession() (*types.Session, error) {
	query := `
	SELECT id, start_time, end_time, duration, project_name, notes, synced_to_sheets, created_at, updated_at
	FROM sessions WHERE end_time IS NULL ORDER BY start_time DESC LIMIT 1`

	row := s.db.QueryRow(query)
	session, err := s.scanSession(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active session
		}
		return nil, err
	}
	return session, nil
}

// GetSessions retrieves sessions based on filter criteria
func (s *SQLiteStorage) GetSessions(filter *types.SessionFilter) ([]*types.Session, error) {
	query := "SELECT id, start_time, end_time, duration, project_name, notes, synced_to_sheets, created_at, updated_at FROM sessions WHERE 1=1"
	args := []interface{}{}

	if filter != nil {
		if filter.StartDate != nil {
			query += " AND start_time >= ?"
			args = append(args, *filter.StartDate)
		}
		if filter.EndDate != nil {
			query += " AND start_time <= ?"
			args = append(args, *filter.EndDate)
		}
		if filter.ProjectName != "" {
			query += " AND project_name = ?"
			args = append(args, filter.ProjectName)
		}
	}

	query += " ORDER BY start_time DESC"

	if filter != nil && filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, types.NewDatabaseErrorWithCause("failed to query sessions", err)
	}
	defer rows.Close()

	var sessions []*types.Session
	for rows.Next() {
		session, err := s.scanSessionFromRows(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, types.NewDatabaseErrorWithCause("error iterating sessions", err)
	}

	return sessions, nil
}

// GetUnsyncedSessions returns sessions that haven't been synced to Google Sheets
func (s *SQLiteStorage) GetUnsyncedSessions() ([]*types.Session, error) {
	query := `
	SELECT id, start_time, end_time, duration, project_name, notes, synced_to_sheets, created_at, updated_at
	FROM sessions WHERE synced_to_sheets = FALSE AND end_time IS NOT NULL
	ORDER BY start_time ASC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, types.NewDatabaseErrorWithCause("failed to query unsynced sessions", err)
	}
	defer rows.Close()

	var sessions []*types.Session
	for rows.Next() {
		session, err := s.scanSessionFromRows(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// MarkSessionsSynced marks multiple sessions as synced
func (s *SQLiteStorage) MarkSessionsSynced(sessionIDs []int64) error {
	if len(sessionIDs) == 0 {
		return nil
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(sessionIDs))
	args := make([]interface{}, len(sessionIDs))
	for i, id := range sessionIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("UPDATE sessions SET synced_to_sheets = TRUE, updated_at = ? WHERE id IN (%s)",
		fmt.Sprintf("%s", placeholders))

	args = append([]interface{}{time.Now()}, args...)

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return types.NewDatabaseErrorWithCause("failed to mark sessions as synced", err)
	}

	return nil
}

// DeleteSession deletes a session by ID
func (s *SQLiteStorage) DeleteSession(id int64) error {
	result, err := s.db.Exec("DELETE FROM sessions WHERE id = ?", id)
	if err != nil {
		return types.NewDatabaseErrorWithCause("failed to delete session", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return types.NewDatabaseErrorWithCause("failed to check delete result", err)
	}

	if rowsAffected == 0 {
		return types.NewNotFoundError("session", fmt.Sprintf("%d", id))
	}

	return nil
}

// GetSessionCount returns the total number of sessions
func (s *SQLiteStorage) GetSessionCount() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&count)
	if err != nil {
		return 0, types.NewDatabaseErrorWithCause("failed to get session count", err)
	}
	return count, nil
}

// GetSessionStats returns basic statistics about sessions
func (s *SQLiteStorage) GetSessionStats() (*types.SessionStatus, error) {
	// Get total sessions count
	var totalSessions int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&totalSessions)
	if err != nil {
		return nil, types.NewDatabaseErrorWithCause("failed to get total sessions", err)
	}

	// Get today's sessions and duration
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	var todaySessions int64
	var todayDuration sql.NullInt64
	err = s.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(duration), 0) 
		FROM sessions 
		WHERE start_time >= ? AND start_time < ?`,
		today, tomorrow).Scan(&todaySessions, &todayDuration)
	if err != nil {
		return nil, types.NewDatabaseErrorWithCause("failed to get today's stats", err)
	}

	// Get active session
	activeSession, err := s.GetActiveSession()
	if err != nil {
		return nil, err
	}

	// Get last sync time from config
	var lastSyncTime time.Time
	err = s.db.QueryRow("SELECT value FROM app_config WHERE key = 'last_sync_time'").Scan(&lastSyncTime)
	if err != nil && err != sql.ErrNoRows {
		return nil, types.NewDatabaseErrorWithCause("failed to get last sync time", err)
	}

	return &types.SessionStatus{
		IsActive:       activeSession != nil,
		CurrentSession: activeSession,
		TotalSessions:  totalSessions,
		TodayDuration:  uint64(todayDuration.Int64),
		TodaySessions:  todaySessions,
		LastSyncTime:   lastSyncTime,
	}, nil
}

// SetConfigValue sets a configuration value
func (s *SQLiteStorage) SetConfigValue(key, value string) error {
	query := `
	INSERT OR REPLACE INTO app_config (key, value, updated_at)
	VALUES (?, ?, ?)`

	_, err := s.db.Exec(query, key, value, time.Now())
	if err != nil {
		return types.NewDatabaseErrorWithCause("failed to set config value", err)
	}
	return nil
}

// GetConfigValue gets a configuration value
func (s *SQLiteStorage) GetConfigValue(key string) (string, error) {
	var value string
	err := s.db.QueryRow("SELECT value FROM app_config WHERE key = ?", key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", types.NewDatabaseErrorWithCause("failed to get config value", err)
	}
	return value, nil
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// scanSession scans a single row into a Session struct
func (s *SQLiteStorage) scanSession(row *sql.Row) (*types.Session, error) {
	var session types.Session
	var endTime sql.NullTime

	err := row.Scan(
		&session.ID,
		&session.StartTime,
		&endTime,
		&session.Duration,
		&session.ProjectName,
		&session.Notes,
		&session.SyncedToSheets,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, types.NewNotFoundError("session", "")
		}
		return nil, types.NewDatabaseErrorWithCause("failed to scan session", err)
	}

	if endTime.Valid {
		session.EndTime = &endTime.Time
	}

	return &session, nil
}

// scanSessionFromRows scans a row from sql.Rows into a Session struct
func (s *SQLiteStorage) scanSessionFromRows(rows *sql.Rows) (*types.Session, error) {
	var session types.Session
	var endTime sql.NullTime

	err := rows.Scan(
		&session.ID,
		&session.StartTime,
		&endTime,
		&session.Duration,
		&session.ProjectName,
		&session.Notes,
		&session.SyncedToSheets,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		return nil, types.NewDatabaseErrorWithCause("failed to scan session from rows", err)
	}

	if endTime.Valid {
		session.EndTime = &endTime.Time
	}

	return &session, nil
}
