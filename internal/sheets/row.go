package sheets

import (
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
	sheet := []any{}

	for _, value := range sessionRecord(s) {
		sheet = append(sheet, value)
	}

	return sheet
}

func SessionToCsvRow(s *session.Session) []string {
	return sessionRecord(s)
}
