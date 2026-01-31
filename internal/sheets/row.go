package sheets

import (
	"slices"
	"time"

	"github.com/vlad/craftie/internal/session"
)

var HEADERS = []any{"Project", "Task", "Date", "Start Time", "End Time", "Duration", "Notes"}

func sessionRecord(s *session.Session) []string {
	endTime := s.EndTime()

	var durationCol string
	if endTime != nil {
		durationCol = endTime.Format(time.TimeOnly)
	} else {
		durationCol = "In progress"
	}

	return []string{
		s.ProjectName,
		s.Task,
		s.StartTime.Format("2006-01-02"),
		s.StartTime.Format(time.TimeOnly),
		durationCol,
		time.Time{}.Add(s.CurrentDuration()).Format(time.TimeOnly),
		s.Notes,
	}
}

func SessionToSheet(s *session.Session) []any {
	record := sessionRecord(s)
	sheet := make([]any, len(record))
	durationIndex := slices.Index(HEADERS, "Duration")

	for i, value := range record {
		if i == durationIndex && s.EndTime() != nil { // Duration column with completed session
			sheet[i] = `=INDIRECT("E"&ROW())-INDIRECT("D"&ROW())`
		} else {
			sheet[i] = value
		}
	}

	return sheet
}

func SessionToCsvRow(s *session.Session) []string {
	return sessionRecord(s)
}
