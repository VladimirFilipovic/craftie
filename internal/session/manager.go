package session

import (
	"time"

	"github.com/vlad/craftie/internal/storage"
	"github.com/vlad/craftie/pkg/types"
)

type Manager struct {
	storage *storage.SQLiteStorage
}

func NewManager(storage *storage.SQLiteStorage) *Manager {
	return &Manager{
		storage: storage,
	}
}

// If there's an active session, it will be stopped automatically
func (m *Manager) StartSession(projectName, notes string) (*types.Session, error) {
	activeSession, err := m.storage.GetActiveSession()
	if err != nil {
		return nil, err
	}

	if activeSession != nil {
		activeSession.Stop()
		if err := m.storage.UpdateSession(activeSession); err != nil {
			return nil, err
		}
	}

	// Create new session
	now := time.Now()
	session := &types.Session{
		StartTime:      now,
		ProjectName:    projectName,
		Notes:          notes,
		SyncedToSheets: false,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Save to database
	if err := m.storage.CreateSession(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (m *Manager) StopSession() (*types.Session, error) {
	activeSession, err := m.storage.GetActiveSession()
	if err != nil {
		return nil, err
	}

	if activeSession == nil {
		return nil, types.ErrNoActiveSession
	}

	// Stop the session
	activeSession.Stop()

	// Update in database
	if err := m.storage.UpdateSession(activeSession); err != nil {
		return nil, err
	}

	return activeSession, nil
}

// GetCurrentSession returns the currently active session, if any
func (m *Manager) GetCurrentSession() (*types.Session, error) {
	return m.storage.GetActiveSession()
}

// TODO move this to some Stats Manager.
// GetStatus returns the current session status
func (m *Manager) GetStatus() (*types.SessionStatus, error) {
	active, err := m.GetCurrentSession()
	if err != nil {
		return nil, err
	}

	total, err := m.storage.GetSessionCount()
	if err != nil {
		return nil, err
	}

	todayStart := time.Now().Truncate(24 * time.Hour)
	todayEnd := todayStart.Add(24 * time.Hour)
	todaySessions, err := m.storage.GetSessions(&types.SessionFilter{
		StartDate: &todayStart,
		EndDate:   &todayEnd,
	})
	if err != nil {
		return nil, err
	}

	todayDuration := uint64(0)
	for _, s := range todaySessions {
		todayDuration += s.GetDuration()
	}

	return &types.SessionStatus{
		IsActive:       active != nil,
		CurrentSession: active,
		TotalSessions:  total,
		TodayDuration:  todayDuration,
		TodaySessions:  int64(len(todaySessions)),
	}, nil
}

func (m *Manager) GetSessions(filter *types.SessionFilter) ([]*types.Session, error) {
	return m.storage.GetSessions(filter)
}

func (m *Manager) GetSessionByID(id int64) (*types.Session, error) {
	return m.storage.GetSessionByID(id)
}

func (m *Manager) UpdateSession(session *types.Session) error {
	return m.storage.UpdateSession(session)
}

func (m *Manager) DeleteSession(id int64) error {
	return m.storage.DeleteSession(id)
}

func (m *Manager) GetDailySummary(startDate, endDate time.Time) ([]*types.SessionSummary, error) {
	// Get all sessions in the date range
	filter := &types.SessionFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	sessions, err := m.storage.GetSessions(filter)
	if err != nil {
		return nil, err
	}

	// Group sessions by date
	dailyMap := make(map[string]*types.SessionSummary)

	for _, session := range sessions {
		dateKey := session.StartTime.Format("2006-01-02")

		summary, exists := dailyMap[dateKey]
		if !exists {
			date, _ := time.Parse("2006-01-02", dateKey)
			summary = &types.SessionSummary{
				Date:          date,
				TotalDuration: 0,
				SessionCount:  0,
				Projects:      []string{},
			}
			dailyMap[dateKey] = summary
		}

		summary.TotalDuration += session.GetDuration()
		summary.SessionCount++

		// Add project to list if not already present
		if session.ProjectName != "" {
			found := false
			for _, project := range summary.Projects {
				if project == session.ProjectName {
					found = true
					break
				}
			}
			if !found {
				summary.Projects = append(summary.Projects, session.ProjectName)
			}
		}
	}

	var summaries []*types.SessionSummary
	for _, summary := range dailyMap {
		summaries = append(summaries, summary)
	}

	// (newest first)
	for i := 0; i < len(summaries)-1; i++ {
		for j := i + 1; j < len(summaries); j++ {
			if summaries[i].Date.Before(summaries[j].Date) {
				summaries[i], summaries[j] = summaries[j], summaries[i]
			}
		}
	}

	return summaries, nil
}

func (m *Manager) AutoSave() error {
	activeSession, err := m.storage.GetActiveSession()
	if err != nil {
		return err
	}

	if activeSession == nil {
		return nil
	}

	activeSession.Duration = activeSession.GetDuration()
	activeSession.UpdatedAt = time.Now()

	return m.storage.UpdateSession(activeSession)
}

func (m *Manager) GetUnsyncedSessions() ([]*types.Session, error) {
	return m.storage.GetUnsyncedSessions()
}

func (m *Manager) MarkSessionsSynced(sessionIDs []int64) error {
	return m.storage.MarkSessionsSynced(sessionIDs)
}
