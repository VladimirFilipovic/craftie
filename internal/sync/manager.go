package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/vlad/craftie/internal/config"
	"github.com/vlad/craftie/internal/sheets"
	"github.com/vlad/craftie/internal/storage"
	"github.com/vlad/craftie/pkg/types"
	sheetsAPI "google.golang.org/api/sheets/v4"
)

// Manager handles synchronization between local storage and Google Sheets
type Manager struct {
	configManager *config.ConfigManager
	sheetsClient  *sheets.SheetsClient
	storage       *storage.SQLiteStorage
}

// NewManager creates a new synchronization manager
func NewManager(configManager *config.ConfigManager, storage *storage.SQLiteStorage) (*Manager, error) {
	sheetsClient, err := sheets.NewSheetsClient(configManager)
	if err != nil {
		return nil, err
	}

	return &Manager{
		configManager: configManager,
		sheetsClient:  sheetsClient,
		storage:       storage,
	}, nil
}

// SyncUnsyncedSessions synchronizes all unsynced sessions to Google Sheets
func (m *Manager) SyncUnsyncedSessions(ctx context.Context) error {
	cfg := m.configManager.Config
	if !cfg.GoogleSheets.Enabled {
		return nil // Skip sync if disabled
	}

	// Get unsynced sessions
	unsyncedSessions, err := m.storage.GetUnsyncedSessions()
	if err != nil {
		return err
	}

	if len(unsyncedSessions) == 0 {
		return nil // Nothing to sync
	}

	// Write sessions to Google Sheets with retry logic
	var lastError error
	for attempt := 0; attempt < cfg.GoogleSheets.RetryAttempts; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-time.After(cfg.GoogleSheets.RetryDelay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := m.sheetsClient.WriteSessions(ctx, unsyncedSessions)
		if err == nil {
			// Success - mark sessions as synced
			sessionIDs := make([]int64, len(unsyncedSessions))
			for i, session := range unsyncedSessions {
				sessionIDs[i] = session.ID
			}
			return m.storage.MarkSessionsSynced(sessionIDs)
		}

		lastError = err
		if attempt == cfg.GoogleSheets.RetryAttempts-1 {
			break
		}
	}

	return fmt.Errorf("failed to sync sessions after %d attempts: %w", cfg.GoogleSheets.RetryAttempts, lastError)
}

// SyncActiveSession synchronizes the currently active session (for progress updates)
func (m *Manager) SyncActiveSession(ctx context.Context) error {
	cfg := m.configManager.Config
	if !cfg.GoogleSheets.Enabled {
		return nil
	}

	activeSession, err := m.storage.GetActiveSession()
	if err != nil {
		return err
	}

	if activeSession == nil {
		return nil // No active session
	}

	// For active sessions, we just update the duration in Google Sheets
	// This is a lightweight operation that doesn't require marking as synced
	return m.sheetsClient.WriteSession(ctx, activeSession)
}

// TestConnection tests the connection to Google Sheets
func (m *Manager) TestConnection(ctx context.Context) error {
	cfg := m.configManager.Config
	if !cfg.GoogleSheets.Enabled {
		return &types.CraftieError{
			Code:    types.ErrCodeConfig,
			Message: "Google Sheets integration is disabled",
		}
	}

	return m.sheetsClient.TestConnection(ctx)
}

// GetSpreadsheetInfo returns information about the configured spreadsheet
func (m *Manager) GetSpreadsheetInfo(ctx context.Context) (*sheetsAPI.Spreadsheet, error) {
	cfg := m.configManager.Config
	if !cfg.GoogleSheets.Enabled {
		return nil, &types.CraftieError{
			Code:    types.ErrCodeConfig,
			Message: "Google Sheets integration is disabled",
		}
	}

	return m.sheetsClient.GetSpreadsheetInfo(ctx)
}

// StartAutoSync starts automatic synchronization in the background
func (m *Manager) StartAutoSync(ctx context.Context, interval time.Duration) <-chan error {
	errChan := make(chan error, 1)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		defer close(errChan)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := m.SyncUnsyncedSessions(ctx); err != nil {
					select {
					case errChan <- err:
					default:
						// Channel is full, drop the error
					}
				}
			}
		}
	}()

	return errChan
}
