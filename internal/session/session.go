package session

import (
	"fmt"
	"time"
)

type Session struct {
	StartTime   time.Time
	EndTime     *time.Time
	ProjectName string
	Notes       string
}

func (s *Session) DurationSec() uint64 {
	return uint64(time.Since(s.StartTime).Seconds())
}

func (s *Session) FormattedDuration() string {
	duration := s.DurationSec()
	hours := duration / 3600
	minutes := (duration % 3600) / 60
	seconds := duration % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

// SetEndTimer parses the duration string, sets the session end time,
// and returns a timer channel that will fire when the session should end.
// Returns nil channel if durationStr is empty.
func (s *Session) SetEndTimer(durationStr string) (<-chan time.Time, error) {
	if durationStr == "" {
		return nil, nil
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return nil, fmt.Errorf("invalid duration format: %w (use format like 2h, 30m, 1h30m)", err)
	}

	endTime := time.Now().Add(duration)
	s.EndTime = &endTime

	fmt.Printf("Session will end automatically in %s (at %s)\n", duration, endTime.Format("15:04:05"))

	return time.After(duration), nil
}
