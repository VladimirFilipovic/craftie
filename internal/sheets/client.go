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

type SheetsClient struct {
	sheetsService *sheets.Service
	config        *config.Config
}

func NewSheetsClient(cfg *config.Config) (*SheetsClient, error) {
	if !cfg.GoogleSheets.Enabled {
		return nil, &types.CraftieError{
			Code:    types.ErrCodeConfig,
			Message: "Google Sheets integration is disabled",
		}
	}

	// TODO: Read this from the console when the app boots
	credsData, err := os.ReadFile(cfg.GoogleSheets.CredentialsFile)
	if err != nil {
		return nil, &types.CraftieError{
			Code:    types.ErrCodeAuth,
			Message: "failed to read service account credentials",
			Cause:   err,
		}
	}

	config, err := google.JWTConfigFromJSON(credsData, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, &types.CraftieError{
			Code:    types.ErrCodeAuth,
			Message: "failed to create JWT config",
			Cause:   err,
		}
	}

	client := config.Client(context.Background())

	srv, err := sheets.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, &types.CraftieError{
			Code:    types.ErrCodeNetwork,
			Message: "failed to create Sheets service",
			Cause:   err,
		}
	}

	return &SheetsClient{
		sheetsService: srv,
		config:        cfg,
	}, nil
}

func (c *SheetsClient) WriteSession(ctx context.Context, session *types.Session) error {
	cfg := c.config

	// Prepare the data to write
	var endTime string
	if session.EndTime != nil {
		endTime = session.EndTime.Format("2006-01-02 15:04:05")
	} else {
		endTime = "ONGOING"
	}

	values := [][]interface{}{
		{
			session.StartTime.Format("2006-01-02 15:04:05"),
			endTime,
			time.Duration(session.Duration) * time.Second,
			session.ProjectName,
			session.Notes,
		},
	}

	// Write to spreadsheet
	_, err := c.sheetsService.Spreadsheets.Values.Update(
		cfg.GoogleSheets.SpreadsheetID,
		fmt.Sprintf("%s!A2:E2", cfg.GoogleSheets.SheetName),
		&sheets.ValueRange{
			Values: values,
		},
	).ValueInputOption("RAW").Do()
	if err != nil {
		return &types.CraftieError{
			Code:    types.ErrCodeNetwork,
			Message: "failed to write session to Google Sheets",
			Cause:   err,
		}
	}

	return nil
}

// WriteSessions writes multiple sessions to Google Sheets
func (c *SheetsClient) WriteSessions(ctx context.Context, sessions []*types.Session) error {
	cfg := c.config

	// Prepare batch data
	values := make([][]interface{}, 0, len(sessions))
	for _, session := range sessions {
		var endTime string
		if session.EndTime != nil {
			endTime = session.EndTime.Format("2006-01-02 15:04:05")
		} else {
			endTime = "ONGOING"
		}

		values = append(values, []interface{}{
			session.StartTime.Format("2006-01-02 15:04:05"),
			endTime,
			time.Duration(session.Duration) * time.Second,
			session.ProjectName,
			session.Notes,
		})
	}

	// Write to spreadsheet in batch
	_, err := c.sheetsService.Spreadsheets.Values.Update(
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
func (c *SheetsClient) TestConnection(ctx context.Context) error {
	cfg := c.config

	// Try to read spreadsheet metadata
	_, err := c.sheetsService.Spreadsheets.Get(cfg.GoogleSheets.SpreadsheetID).Do()
	if err != nil {
		return &types.CraftieError{
			Code:    types.ErrCodeNetwork,
			Message: "failed to connect to Google Sheets",
			Cause:   err,
		}
	}

	return nil
}

func (c *SheetsClient) GetSpreadsheetInfo(ctx context.Context) (*sheets.Spreadsheet, error) {
	cfg := c.config

	spreadsheet, err := c.sheetsService.Spreadsheets.Get(cfg.GoogleSheets.SpreadsheetID).Do()
	if err != nil {
		return nil, &types.CraftieError{
			Code:    types.ErrCodeNetwork,
			Message: "failed to get spreadsheet info",
			Cause:   err,
		}
	}

	return spreadsheet, nil
}
