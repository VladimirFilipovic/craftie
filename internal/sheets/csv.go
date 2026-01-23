package sheets

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vlad/craftie/internal/session"
)

// CsvSyncState tracks the CSV file for syncing
type CsvSyncState struct {
	FilePath  string
	RowOffset int64 // byte offset where the row starts
}

// InitCsvRow creates the initial row for an in-progress session
func InitCsvRow(filePath string, session *session.Session) (*CsvSyncState, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	// Read existing rows
	reader := csv.NewReader(file)
	existingRows, _ := reader.ReadAll()

	writer := csv.NewWriter(file)

	// Add headers if empty
	if len(existingRows) == 0 {
		headers := []string{"Project", "Task", "Date", "Start Time", "End Time", "Duration", "Notes"}
		if err := writer.Write(headers); err != nil {
			return nil, fmt.Errorf("failed to write CSV headers: %w", err)
		}
		writer.Flush()
	}

	// Get current file size - this is where our row will start
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	rowOffset := stat.Size()

	record := []string{
		session.ProjectName,
		session.Task,
		session.StartTime.Format(strings.ReplaceAll(time.DateOnly, ":", "/")),
		session.StartTime.Format(time.TimeOnly),
		"In Progress",
		time.Time{}.Add(session.CurrentDuration()).Format(time.TimeOnly),
		session.Notes,
	}

	if err := writer.Write(record); err != nil {
		return nil, fmt.Errorf("failed to write CSV record: %w", err)
	}
	writer.Flush()

	return &CsvSyncState{FilePath: filePath, RowOffset: rowOffset}, nil
}

// SyncCsvRow updates the row at RowOffset with current session data
func SyncCsvRow(state *CsvSyncState, session *session.Session) error {
	file, err := os.OpenFile(state.FilePath, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	// Truncate to row start, then seek there to write
	if err := file.Truncate(state.RowOffset); err != nil {
		return fmt.Errorf("failed to truncate: %w", err)
	}

	var endTimeStr string
	if endTime := session.EndTime(); endTime != nil {
		endTimeStr = endTime.Format(time.TimeOnly)
	} else {
		endTimeStr = "In Progress"
	}

	record := []string{
		session.ProjectName,
		session.Task,
		session.StartTime.Format(strings.ReplaceAll(time.DateOnly, ":", "/")),
		session.StartTime.Format(time.TimeOnly),
		endTimeStr,
		time.Time{}.Add(session.CurrentDuration()).Format(time.TimeOnly),
		session.Notes,
	}

	writer := csv.NewWriter(file)
	if err := writer.Write(record); err != nil {
		return fmt.Errorf("failed to write CSV record: %w", err)
	}
	writer.Flush()

	return nil
}
