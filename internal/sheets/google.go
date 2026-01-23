package sheets

import (
	"context"
	"fmt"
	"time"

	"github.com/vlad/craftie/internal/config"
	"github.com/vlad/craftie/internal/session"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// NewSheetsClient creates a Google Sheets service client.
func NewSheetsClient(ctx context.Context, credentialsHelper string) (*sheets.Service, error) {
	credentials, err := GetCredentials(credentialsHelper)
	if err != nil {
		return nil, err
	}

	jwtConfig, err := google.JWTConfigFromJSON(credentials, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	httpClient := jwtConfig.Client(ctx)
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	return srv, nil
}

type GoogleSheetsParams struct {
	Srv     *sheets.Service
	Cfg     config.GoogleSheetsConfig
	Session *session.Session
}

// SyncState tracks the row number for updating an in-progress session
type SyncState struct {
	RowNumber int64
}

// InitRow creates the initial row for an in-progress session
func InitRow(ctx context.Context, p GoogleSheetsParams) (*SyncState, error) {
	quotedSheetName := fmt.Sprintf("'%s'", p.Cfg.SheetName)

	// Check if sheet has headers
	readRange := fmt.Sprintf("%s!A1:G1", quotedSheetName)
	resp, err := p.Srv.Spreadsheets.Values.Get(p.Cfg.SpreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet headers: %w", err)
	}

	// If sheet is empty, add headers
	if len(resp.Values) == 0 {
		headers := []any{"Project", "Task", "Date", "Start Time", "End Time", "Duration", "Notes"}
		headerRange := fmt.Sprintf("%s!A1:G1", quotedSheetName)
		headerValueRange := &sheets.ValueRange{
			Values: [][]any{headers},
		}
		_, err = p.Srv.Spreadsheets.Values.Update(p.Cfg.SpreadsheetID, headerRange, headerValueRange).
			ValueInputOption("USER_ENTERED").Do()
		if err != nil {
			return nil, fmt.Errorf("failed to write headers: %w", err)
		}
	}

	// Create initial row with in-progress marker
	row := []any{
		p.Session.ProjectName,
		p.Session.Task,
		p.Session.StartTime.Format("2006-01-02"),
		p.Session.StartTime.Format(time.TimeOnly),
		"In Progress",
		time.Time{}.Add(p.Session.CurrentDuration()).Format(time.TimeOnly),
		p.Session.Notes,
	}

	appendRange := fmt.Sprintf("%s!A:G", quotedSheetName)
	valueRange := &sheets.ValueRange{
		Values: [][]any{row},
	}

	appendResp, err := p.Srv.Spreadsheets.Values.Append(p.Cfg.SpreadsheetID, appendRange, valueRange).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to append row: %w", err)
	}

	// Parse the updated range to get the row number
	// Format is like 'Sheet Name'!A5:F5
	var rowNum int64
	fmt.Sscanf(appendResp.Updates.UpdatedRange, "%*[^0-9]%d", &rowNum)

	return &SyncState{RowNumber: rowNum}, nil
}

// SyncGoogleSheetsRow updates an existing row with current session duration
func SyncGoogleSheetsRow(ctx context.Context, p GoogleSheetsParams, state *SyncState) error {
	quotedSheetName := fmt.Sprintf("'%s'", p.Cfg.SheetName)

	endTime := p.Session.EndTime()

	var durationCol string
	if endTime != nil {
		durationCol = endTime.Format(time.TimeOnly)
	}

	if endTime == nil {
		durationCol = "In progress"
	}

	row := []any{
		p.Session.ProjectName,
		p.Session.Task,
		p.Session.StartTime.Format("2006-01-02"),
		p.Session.StartTime.Format(time.TimeOnly),
		durationCol,
		time.Time{}.Add(p.Session.CurrentDuration()).Format(time.TimeOnly),
		p.Session.Notes,
	}

	updateRange := fmt.Sprintf("%s!A%d:G%d", quotedSheetName, state.RowNumber, state.RowNumber)
	valueRange := &sheets.ValueRange{
		Values: [][]any{row},
	}

	_, err := p.Srv.Spreadsheets.Values.Update(p.Cfg.SpreadsheetID, updateRange, valueRange).
		ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return fmt.Errorf("failed to update row: %w", err)
	}

	return nil
}
