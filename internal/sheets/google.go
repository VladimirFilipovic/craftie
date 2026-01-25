package sheets

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

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
	RowNumber int
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
		headerRange := fmt.Sprintf("%s!A1:G1", quotedSheetName)
		headerValueRange := &sheets.ValueRange{
			Values: [][]any{HEADERS},
		}
		_, err = p.Srv.Spreadsheets.Values.Update(p.Cfg.SpreadsheetID, headerRange, headerValueRange).
			ValueInputOption("USER_ENTERED").Do()
		if err != nil {
			return nil, fmt.Errorf("failed to write headers: %w", err)
		}
	}

	appendRange := fmt.Sprintf("%s!A:G", quotedSheetName)
	valueRange := &sheets.ValueRange{
		Values: [][]any{SessionToSheet(p.Session)},
	}

	appendResp, err := p.Srv.Spreadsheets.Values.Append(p.Cfg.SpreadsheetID, appendRange, valueRange).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to append row: %w", err)
	}

	var rowNum = 0
	re := regexp.MustCompile(`![A-Z]+(\d+)`)
	matches := re.FindStringSubmatch(appendResp.Updates.UpdatedRange)
	if len(matches) > 1 {
		rowNum, _ = strconv.Atoi(matches[1])
	}

	return &SyncState{RowNumber: rowNum}, nil
}

// SyncGoogleSheetsRow updates an existing row with current session duration
func SyncGoogleSheetsRow(ctx context.Context, p GoogleSheetsParams, state *SyncState) error {
	quotedSheetName := fmt.Sprintf("'%s'", p.Cfg.SheetName)

	updateRange := fmt.Sprintf("%s!A%d:G%d", quotedSheetName, state.RowNumber, state.RowNumber)
	valueRange := &sheets.ValueRange{
		Values: [][]any{SessionToSheet(p.Session)},
	}

	_, err := p.Srv.Spreadsheets.Values.Update(p.Cfg.SpreadsheetID, updateRange, valueRange).
		ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return fmt.Errorf("failed to update row: %w", err)
	}

	return nil
}
