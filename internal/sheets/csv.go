package sheets

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vlad/craftie/internal/session"
)

// CsvRecorder implements session.Recorder for CSV files
type CsvRecorder struct {
	filePath    string
	rowOffset   int64
	initialized bool
}

// NewCsvRecorder creates a new CSV recorder
func NewCsvRecorder(filePath string) *CsvRecorder {
	return &CsvRecorder{filePath: filePath}
}

// Record saves the session to CSV, creating the row on first call
func (r *CsvRecorder) Record(sess *session.Session) error {
	if r.initialized {
		return r.update(sess)
	}
	return r.init(sess)
}

func (r *CsvRecorder) init(sess *session.Session) error {
	dir := filepath.Dir(r.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.OpenFile(r.filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	existingRows, _ := reader.ReadAll()

	writer := csv.NewWriter(file)

	if len(existingRows) == 0 {
		headers := []string{"Project", "Task", "Date", "Start Time", "End Time", "Duration", "Notes"}
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("failed to write CSV headers: %w", err)
		}
		writer.Flush()
	}

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	r.rowOffset = stat.Size()

	record := SessionToCsvRow(sess)
	if err := writer.Write(record); err != nil {
		return fmt.Errorf("failed to write CSV record: %w", err)
	}
	writer.Flush()

	r.initialized = true
	fmt.Printf("Session row created in CSV: %s\n", r.filePath)
	return nil
}

func (r *CsvRecorder) update(sess *session.Session) error {
	file, err := os.OpenFile(r.filePath, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	if err := file.Truncate(r.rowOffset); err != nil {
		return fmt.Errorf("failed to truncate: %w", err)
	}

	record := SessionToCsvRow(sess)
	writer := csv.NewWriter(file)
	if err := writer.Write(record); err != nil {
		return fmt.Errorf("failed to write CSV record: %w", err)
	}
	writer.Flush()

	return nil
}
