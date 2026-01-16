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

// SaveToGoogleSheets appends a session record to a Google Sheet
func SaveToGoogleSheets(ctx context.Context, srv *sheets.Service, cfg config.GoogleSheetsConfig, session *session.Session) error {

	duration, err := session.Duration()
	if err != nil {
		return fmt.Errorf("failed to calculate duration: %w", err)
	}

	endTime := session.EndTime()
	if endTime == nil {
		return fmt.Errorf("session has no end time")
	}

	row := []interface{}{
		session.ProjectName,
		session.StartTime.Format("2006-01-02"),
		session.StartTime.Format(time.TimeOnly),
		endTime.Format(time.TimeOnly),
		time.Time{}.Add(duration).Format(time.TimeOnly),
		session.Notes,
	}

	// Check if sheet has headers
	// Quote sheet name to handle special characters like spaces and hyphens
	quotedSheetName := fmt.Sprintf("'%s'", cfg.SheetName)
	readRange := fmt.Sprintf("%s!A1:F1", quotedSheetName)
	resp, err := srv.Spreadsheets.Values.Get(cfg.SpreadsheetID, readRange).Do()
	if err != nil {
		return fmt.Errorf("failed to read sheet headers: %w", err)
	}

	// If sheet is empty, add headers
	if len(resp.Values) == 0 {
		headers := []interface{}{"Project", "Date", "Start Time", "End Time", "Duration", "Notes"}
		headerRange := fmt.Sprintf("%s!A1:F1", quotedSheetName)
		headerValueRange := &sheets.ValueRange{
			Values: [][]interface{}{headers},
		}
		_, err = srv.Spreadsheets.Values.Update(cfg.SpreadsheetID, headerRange, headerValueRange).
			ValueInputOption("USER_ENTERED").Do()
		if err != nil {
			return fmt.Errorf("failed to write headers: %w", err)
		}
	}

	appendRange := fmt.Sprintf("%s!A:F", quotedSheetName)
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{row},
	}

	_, err = srv.Spreadsheets.Values.Append(cfg.SpreadsheetID, appendRange, valueRange).
		ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return fmt.Errorf("failed to append row: %w", err)
	}

	return nil
}
