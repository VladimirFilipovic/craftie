package sheets

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/vlad/craftie/internal/config"
	"github.com/vlad/craftie/internal/keyring"
	"github.com/vlad/craftie/internal/session"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// GoogleSheetsRecorder implements session.Recorder for Google Sheets
type GoogleSheetsRecorder struct {
	client      *sheets.Service
	cfg         config.GoogleSheetsConfig
	rowNumber   int
	initialized bool
}

// NewGoogleSheetsRecorder creates a new Google Sheets recorder
func NewGoogleSheetsRecorder(ctx context.Context, cfg config.GoogleSheetsConfig) (*GoogleSheetsRecorder, error) {
	keyringSession, err := keyring.Open()
	if err != nil {
		return nil, err
	}
	defer keyringSession.Close()

	credentials, err := GetCredentials(cfg.CredentialsHelper, keyringSession)
	if err != nil {
		return nil, err
	}

	jwtConfig, err := google.JWTConfigFromJSON(credentials, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	httpClient := jwtConfig.Client(ctx)
	client, err := sheets.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	return &GoogleSheetsRecorder{
		client: client,
		cfg:    cfg,
	}, nil
}

// Record saves the session to Google Sheets, creating the row on first call
func (r *GoogleSheetsRecorder) Record(sess *session.Session) error {
	if r.initialized {
		return r.update(sess)
	}
	return r.init(sess)
}

func (r *GoogleSheetsRecorder) init(sess *session.Session) error {
	quotedSheetName := fmt.Sprintf("'%s'", r.cfg.SheetName)

	// Check if sheet has headers
	readRange := fmt.Sprintf("%s!A1:G1", quotedSheetName)
	resp, err := r.client.Spreadsheets.Values.Get(r.cfg.SpreadsheetID, readRange).Do()
	if err != nil {
		return fmt.Errorf("failed to read sheet headers: %w", err)
	}

	// If sheet is empty, add headers
	if len(resp.Values) == 0 {
		headerRange := fmt.Sprintf("%s!A1:G1", quotedSheetName)
		headerValueRange := &sheets.ValueRange{
			Values: [][]any{HEADERS},
		}
		_, err = r.client.Spreadsheets.Values.Update(r.cfg.SpreadsheetID, headerRange, headerValueRange).
			ValueInputOption("USER_ENTERED").Do()
		if err != nil {
			return fmt.Errorf("failed to write headers: %w", err)
		}
	}

	appendRange := fmt.Sprintf("%s!A:G", quotedSheetName)
	valueRange := &sheets.ValueRange{
		Values: [][]any{SessionToSheet(sess)},
	}

	appendResp, err := r.client.Spreadsheets.Values.Append(r.cfg.SpreadsheetID, appendRange, valueRange).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Do()
	if err != nil {
		return fmt.Errorf("failed to append row: %w", err)
	}

	re := regexp.MustCompile(`![A-Z]+(\d+)`)
	matches := re.FindStringSubmatch(appendResp.Updates.UpdatedRange)
	if len(matches) > 1 {
		r.rowNumber, _ = strconv.Atoi(matches[1])
	}

	r.initialized = true
	fmt.Printf("Session row created in Google Sheets (row %d)\n", r.rowNumber)
	return nil
}

func (r *GoogleSheetsRecorder) update(sess *session.Session) error {
	quotedSheetName := fmt.Sprintf("'%s'", r.cfg.SheetName)

	updateRange := fmt.Sprintf("%s!A%d:G%d", quotedSheetName, r.rowNumber, r.rowNumber)
	valueRange := &sheets.ValueRange{
		Values: [][]any{SessionToSheet(sess)},
	}

	_, err := r.client.Spreadsheets.Values.Update(r.cfg.SpreadsheetID, updateRange, valueRange).
		ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return fmt.Errorf("failed to update row: %w", err)
	}

	return nil
}
