package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	sql "github.com/jmoiron/sqlx"
	"github.com/vlad/craftie/internal/path"
	"github.com/vlad/craftie/pkg/types"
	_ "modernc.org/sqlite"
)

type SQLiteStorage struct {
	*sql.DB
	path string
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	dbPath, err := path.ExpandPathWithHome(dbPath)

	if err != nil {
		return nil, types.NewDatabaseErrorWithCause("failed to get home directory", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, types.NewDatabaseErrorWithCause("failed to create database directory", err)
	}

	db, err := sql.Connect("sqlite", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, types.NewDatabaseErrorWithCause("failed to open database", err)
	}

	storage := &SQLiteStorage{
		db,
		dbPath,
	}

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

	if _, err := s.Exec(createSessionsTable); err != nil {
		return types.NewDatabaseErrorWithCause("failed to create sessions table", err)
	}

	// Create indexes for performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_sessions_project ON sessions(project_name);",
	}

	for _, indexSQL := range indexes {
		if _, err := s.Exec(indexSQL); err != nil {
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

	if _, err := s.Exec(createConfigTable); err != nil {
		return types.NewDatabaseErrorWithCause("failed to create config table", err)
	}

	return nil
}

func (s *SQLiteStorage) CreateSession(session *types.Session) error {
	insertStatement := `
	INSERT INTO sessions (start_time, end_time, duration, project_name, notes, synced_to_sheets, created_at, updated_at)
	VALUES (:start_time, :end_time, :duration, :project_name, :notes, :synced_to_sheets, :created_at, :updated_at)`

	result, err := s.NamedExec(insertStatement,
		session,
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

	result, err := s.Exec(query,
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

func (s *SQLiteStorage) GetSessionByID(id int64) (*types.Session, error) {
	var session types.Session
	err := s.Get(&session, "SELECT id, start_time, end_time, duration, project_name, notes, synced_to_sheets, created_at, updated_at FROM sessions WHERE id = ?", id)
	if err != nil {
		// Check if it's a "no rows" error by trying to query again
		var count int
		err2 := s.Get(&count, "SELECT COUNT(*) FROM sessions WHERE id = ?", id)
		if err2 == nil && count == 0 {
			return nil, nil // No rows found
		}
		return nil, types.NewDatabaseErrorWithCause("failed to get session by ID", err)
	}
	return &session, nil
}

func (s *SQLiteStorage) GetActiveSession() (*types.Session, error) {
	var session types.Session
	err := s.Get(&session, "SELECT id, start_time, end_time, duration, project_name, notes, synced_to_sheets, created_at, updated_at FROM sessions WHERE end_time IS NULL ORDER BY start_time DESC LIMIT 1")
	if err != nil {
		// Check if it's a "no rows" error by trying to query again
		var count int
		err2 := s.Get(&count, "SELECT COUNT(*) FROM sessions WHERE end_time IS NULL")
		if err2 == nil && count == 0 {
			return nil, nil // No active session
		}
		return nil, types.NewDatabaseErrorWithCause("failed to get active session", err)
	}
	return &session, nil
}

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

	var sessions []*types.Session
	err := s.Select(&sessions, query, args...)
	if err != nil {
		return nil, types.NewDatabaseErrorWithCause("failed to query sessions", err)
	}

	return sessions, nil
}

func (s *SQLiteStorage) GetUnsyncedSessions() ([]*types.Session, error) {
	query := `
	SELECT id, start_time, end_time, duration, project_name, notes, synced_to_sheets, created_at, updated_at
	FROM sessions WHERE synced_to_sheets = FALSE AND end_time IS NOT NULL
	ORDER BY start_time ASC`

	var sessions []*types.Session
	err := s.Select(&sessions, query)
	if err != nil {
		return nil, types.NewDatabaseErrorWithCause("failed to query unsynced sessions", err)
	}

	return sessions, nil
}

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

	_, err := s.Exec(query, args...)
	if err != nil {
		return types.NewDatabaseErrorWithCause("failed to mark sessions as synced", err)
	}

	return nil
}

func (s *SQLiteStorage) DeleteSession(id int64) error {
	result, err := s.Exec("DELETE FROM sessions WHERE id = ?", id)
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

func (s *SQLiteStorage) GetSessionCount() (int64, error) {
	var count int64
	err := s.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&count)
	if err != nil {
		return 0, types.NewDatabaseErrorWithCause("failed to get session count", err)
	}
	return count, nil
}

func (s *SQLiteStorage) GetSessionStats() (*types.SessionStatus, error) {
	// Get total sessions count
	var totalSessions int64
	err := s.Get(&totalSessions, "SELECT COUNT(*) FROM sessions")
	if err != nil {
		return nil, types.NewDatabaseErrorWithCause("failed to get total sessions", err)
	}

	// Get today's sessions and duration
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	var todayDuration int64
	err = s.Get(&todayDuration, `
		SELECT COALESCE(SUM(duration), 0)
		FROM sessions
		WHERE start_time >= ? AND start_time < ?`,
		today, tomorrow)
	if err != nil {
		return nil, types.NewDatabaseErrorWithCause("failed to get today's stats", err)
	}

	// Get today's sessions count
	var todaySessions int64
	err = s.Get(&todaySessions, `
		SELECT COUNT(*)
		FROM sessions
		WHERE start_time >= ? AND start_time < ?`,
		today, tomorrow)
	if err != nil {
		return nil, types.NewDatabaseErrorWithCause("failed to get today's sessions count", err)
	}

	// Get active session
	activeSession, err := s.GetActiveSession()
	if err != nil {
		return nil, err
	}

	// Get last sync time from config
	var lastSyncTime time.Time
	err = s.Get(&lastSyncTime, "SELECT value FROM app_config WHERE key = 'last_sync_time'")
	if err != nil {
		// Check if it's a "no rows" error by trying to query again
		var count int
		err2 := s.Get(&count, "SELECT COUNT(*) FROM app_config WHERE key = 'last_sync_time'")
		if err2 == nil && count == 0 {
			lastSyncTime = time.Time{} // Zero time if not found
		} else {
			return nil, types.NewDatabaseErrorWithCause("failed to get last sync time", err)
		}
	}

	return &types.SessionStatus{
		IsActive:       activeSession != nil,
		CurrentSession: activeSession,
		TotalSessions:  totalSessions,
		TodayDuration:  uint64(todayDuration),
		TodaySessions:  todaySessions,
		LastSyncTime:   lastSyncTime,
	}, nil
}

func (s *SQLiteStorage) SetConfigValue(key, value string) error {
	query := `
	INSERT OR REPLACE INTO app_config (key, value, updated_at)
	VALUES (?, ?, ?)`

	_, err := s.Exec(query, key, value, time.Now())
	if err != nil {
		return types.NewDatabaseErrorWithCause("failed to set config value", err)
	}
	return nil
}

func (s *SQLiteStorage) GetConfigValue(key string) (string, error) {
	var value string
	err := s.Get(&value, "SELECT value FROM app_config WHERE key = ?", key)
	if err != nil {
		// Check if it's a "no rows" error by trying to query again
		var count int
		err2 := s.Get(&count, "SELECT COUNT(*) FROM app_config WHERE key = ?", key)
		if err2 == nil && count == 0 {
			return "", nil // Key doesn't exist
		}
		return "", types.NewDatabaseErrorWithCause("failed to get config value", err)
	}
	return value, nil
}

func (s *SQLiteStorage) Close() error {
	if s.Stats().OpenConnections > 0 {
		return s.Close()
	}
	return nil
}
