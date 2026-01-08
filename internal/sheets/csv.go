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

func SaveToCsv(filePath string, session *session.Session) error {
	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if file exists before opening
	_, err := os.Stat(filePath)
	fileExists := !os.IsNotExist(err)

	// Open file in append mode
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers if file is new
	if !fileExists {
		headers := []string{"Project", "Date", "Start Time", "End Time", "Duration", "Notes"}
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("failed to write CSV headers: %w", err)
		}
	}

	// Calculate duration
	duration, err := session.Duration()
	if err != nil {
		return err
	}

	// Get end time
	endTime := session.EndTime()
	if endTime == nil {
		return fmt.Errorf("session has no end time")
	}

	// Prepare record
	record := []string{
		session.ProjectName,
		session.StartTime.Format(strings.Replace(time.DateOnly, ":", "/", -1)),
		session.StartTime.Format(time.TimeOnly),
		endTime.Format(time.TimeOnly),
		time.Time{}.Add(duration).Format(time.TimeOnly),
		session.Notes,
	}

	// Write record
	if err := writer.Write(record); err != nil {
		return fmt.Errorf("failed to write CSV record: %w", err)
	}

	return nil
}
