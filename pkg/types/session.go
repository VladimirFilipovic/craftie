package types

import (
	"time"
)

// Session represents a craft time tracking session
type Session struct {
	ID             int64      `json:"id" db:"id"`
	StartTime      time.Time  `json:"start_time" db:"start_time"`
	EndTime        *time.Time `json:"end_time,omitempty" db:"end_time"`
	Duration       uint64     `json:"duration" db:"duration"` // seconds
	ProjectName    string     `json:"project_name" db:"project_name"`
	Notes          string     `json:"notes" db:"notes"`
	SyncedToSheets bool       `json:"synced_to_sheets" db:"synced_to_sheets"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// IsActive returns true if the session is currently active (no end time)
func (s *Session) IsActive() bool {
	return s.EndTime == nil
}

// GetDuration returns the session duration in seconds
// If session is active, calculates duration from start time to now
func (s *Session) GetDuration() uint64 {
	if s.EndTime != nil {
		return s.Duration
	}
	// Session is active, calculate current duration
	return uint64(time.Since(s.StartTime).Seconds())
}

// Stop ends the session and calculates final duration
func (s *Session) Stop() {
	now := time.Now()
	s.EndTime = &now
	s.Duration = uint64(now.Sub(s.StartTime).Seconds())
	s.UpdatedAt = now
}

// SessionStatus represents the current status of the tracking system
type SessionStatus struct {
	IsActive       bool      `json:"is_active"`
	CurrentSession *Session  `json:"current_session,omitempty"`
	TotalSessions  int64     `json:"total_sessions"`
	TodayDuration  uint64    `json:"today_duration"` // seconds
	TodaySessions  int64     `json:"today_sessions"`
	LastSyncTime   time.Time `json:"last_sync_time"`
}

// SessionSummary represents aggregated session data
type SessionSummary struct {
	Date          time.Time `json:"date"`
	TotalDuration uint64    `json:"total_duration"` // seconds
	SessionCount  int64     `json:"session_count"`
	Projects      []string  `json:"projects"`
}

// SessionFilter represents filtering options for session queries
type SessionFilter struct {
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	ProjectName string     `json:"project_name,omitempty"`
	Limit       int        `json:"limit,omitempty"`
	Offset      int        `json:"offset,omitempty"`
}

// SessionManager interface defines session management operations
type SessionManager interface {
	// StartSession creates and starts a new session
	// If there's an active session, it will be stopped automatically
	StartSession(projectName, notes string) (*Session, error)

	// StopSession stops the current active session
	StopSession() (*Session, error)

	// GetCurrentSession returns the currently active session, if any
	GetCurrentSession() (*Session, error)

	// GetStatus returns the current system status
	GetStatus() (*SessionStatus, error)

	// GetSessions returns sessions based on filter criteria
	GetSessions(filter *SessionFilter) ([]*Session, error)

	// GetSessionByID returns a specific session by ID
	GetSessionByID(id int64) (*Session, error)

	// UpdateSession updates an existing session
	UpdateSession(session *Session) error

	// DeleteSession deletes a session by ID
	DeleteSession(id int64) error

	// GetDailySummary returns daily summary for a date range
	GetDailySummary(startDate, endDate time.Time) ([]*SessionSummary, error)

	// GetSessionCount returns the total number of sessions
	GetSessionCount() (int64, error)
}
