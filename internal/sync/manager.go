package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/vlad/craftie/internal/config"
	"github.com/vlad/craftie/internal/sheets"
	"github.com/vlad/craftie/pkg/types"
	sheetsAPI "google.golang.org/api/sheets/v4"
)

// Manager handles synchronization to Google Sheets
type Manager struct {
	config       *config.Config
	sheetsClient *sheets.SheetsClient
}

// NewManager creates a new synchronization manager
func NewManager(cfg *config.Config) (*Manager, error) {
	sheetsClient, err := sheets.NewSheetsClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Manager{
		config:       cfg,
		sheetsClient: sheetsClient,
	}, nil
}

// SyncSessions synchronizes sessions to Google Sheets
func (m *Manager) SyncSessions(ctx context.Context, sessions []*types.Session) error {
	if !m.config.GoogleSheets.Enabled {
		return nil // Skip sync if disabled
	}

	if len(sessions) == 0 {
		return nil // Nothing to sync
	}

	// Write sessions to Google Sheets with retry logic
	const maxRetries = 3
	const retryDelay = 2 * time.Second

	var lastError error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-time.After(retryDelay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := m.sheetsClient.WriteSessions(ctx, sessions)
		if err == nil {
			return nil
		}

		lastError = err
		if attempt == maxRetries-1 {
			break
		}
	}

	return fmt.Errorf("failed to sync sessions after %d attempts: %w", maxRetries, lastError)
}

// SyncSession synchronizes a single session to Google Sheets
func (m *Manager) SyncSession(ctx context.Context, session *types.Session) error {
	if !m.config.GoogleSheets.Enabled {
		return nil
	}

	if session == nil {
		return nil
	}

	return m.sheetsClient.WriteSession(ctx, session)
}

// TestConnection tests the connection to Google Sheets
func (m *Manager) TestConnection(ctx context.Context) error {
	if !m.config.GoogleSheets.Enabled {
		return &types.CraftieError{
			Code:    types.ErrCodeConfig,
			Message: "Google Sheets integration is disabled",
		}
	}

	return m.sheetsClient.TestConnection(ctx)
}

// GetSpreadsheetInfo returns information about the configured spreadsheet
func (m *Manager) GetSpreadsheetInfo(ctx context.Context) (*sheetsAPI.Spreadsheet, error) {
	if !m.config.GoogleSheets.Enabled {
		return nil, &types.CraftieError{
			Code:    types.ErrCodeConfig,
			Message: "Google Sheets integration is disabled",
		}
	}

	return m.sheetsClient.GetSpreadsheetInfo(ctx)
}
