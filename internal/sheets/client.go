package sheets

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"github.com/vlad/craftie/internal/config"
	"github.com/vlad/craftie/pkg/types"
)

// Client handles Google Sheets API operations
type Client struct {
	service *sheets.Service
	config  *config.Manager
}

// NewClient creates a new Google Sheets client
func NewClient(configManager *config.Manager) (*Client, error) {
	cfg := configManager.GetConfig()
	if !cfg.GoogleSheets.Enabled {
		return nil, &types.CraftieError{
			Code:    types.ErrCodeConfig,
			Message: "Google Sheets integration is disabled",
		}
	}

	// Read service account credentials
	credsData, err := os.ReadFile(cfg.GoogleSheets.CredentialsFile)
	if err != nil {
		return nil, &types.CraftieError{
			Code:    types.ErrCodeAuth,
			Message: "failed to read service account credentials",
			Cause:   err,
		}
	}

	// Create OAuth2 config
	config, err := google.JWTConfigFromJSON(credsData, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, &types.CraftieError{
			Code:    types.ErrCodeAuth,
			Message: "failed to create JWT config",
			Cause:   err,
		}
	}

	// Create HTTP client
	client := config.Client(context.Background())

	// Create Sheets service
	srv, err := sheets.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, &types.CraftieError{
			Code:    types.ErrCodeNetwork,
			Message: "failed to create Sheets service",
			Cause:   err,
		}
	}

	return &Client{
		service: srv,
		config:  configManager,
	}, nil
}

// WriteSession writes a session to Google Sheets
func (c *Client) WriteSession(ctx context.Context, session *types.Session) error {
	cfg := c.config.GetConfig()

	// Prepare the data to write
	values := [][]interface{}{
		{
			session.StartTime.Format("2006-01-02 15:04:05"),
			session.EndTime.Format("2006-01-02 15:04:05"),
			time.Duration(session.Duration) * time.Second,
			session.ProjectName,
			session.Notes,
		},
	}

	// Write to spreadsheet
	_, err := c.service.Spreadsheets.Values.Update(
		cfg.GoogleSheets.SpreadsheetID,
		fmt.Sprintf("%s!A2:E2", cfg.GoogleSheets.SheetName),
		&sheets.ValueRange{
			Values: values,
		},
	).ValueInputOption("RAW").Do()
	if err != nil {
		return &types.CraftieError{
			Code:    types.ErrCodeConfig,
			Message: "Google Sheets integration is disabled",
			Cause:   err,
		}
	}

	return nil
}

// WriteSessions writes multiple sessions to Google Sheets
func (c *Client) WriteSessions(ctx context.Context, sessions []*types.Session) error {
	cfg := c.config.GetConfig()

	// Prepare batch data
	values := make([][]interface{}, 0, len(sessions))
	for _, session := range sessions {
		values = append(values, []interface{}{
			session.StartTime.Format("2006-01-02 15:04:05"),
			session.EndTime.Format("2006-01-02 15:04:05"),
			time.Duration(session.Duration) * time.Second,
			session.ProjectName,
			session.Notes,
		})
	}

	// Write to spreadsheet in batch
	_, err := c.service.Spreadsheets.Values.Update(
		cfg.GoogleSheets.SpreadsheetID,
		fmt.Sprintf("%s!A2:E%d", cfg.GoogleSheets.SheetName, len(values)+1),
		&sheets.ValueRange{
			Values: values,
		},
	).ValueInputOption("RAW").Do()
	if err != nil {
		return &types.CraftieError{
			Code:    types.ErrCodeNetwork,
			Message: "failed to write sessions to Google Sheets",
			Cause:   err,
		}
	}

	return nil
}

// TestConnection tests the connection to Google Sheets
func (c *Client) TestConnection(ctx context.Context) error {
	cfg := c.config.GetConfig()

	// Try to read spreadsheet metadata
	_, err := c.service.Spreadsheets.Get(cfg.GoogleSheets.SpreadsheetID).Do()
	if err != nil {
		return &types.CraftieError{
			Code:    types.ErrCodeNetwork,
			Message: "failed to connect to Google Sheets",
			Cause:   err,
		}
	}

	return nil
}

// GetSpreadsheetInfo returns information about the configured spreadsheet
func (c *Client) GetSpreadsheetInfo(ctx context.Context) (*sheets.Spreadsheet, error) {
	cfg := c.config.GetConfig()

	spreadsheet, err := c.service.Spreadsheets.Get(cfg.GoogleSheets.SpreadsheetID).Do()
	if err != nil {
		return nil, &types.CraftieError{
			Code:    types.ErrCodeNetwork,
			Message: "failed to get spreadsheet info",
			Cause:   err,
		}
	}

	return spreadsheet, nil
}
