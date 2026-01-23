package session

import (
	"fmt"
	"time"
)

type Session struct {
	StartTime   time.Time
	endTime     *time.Time
	ProjectName string
	Task        string
	Notes       string
}

// Duration returns the duration from start until now (for in-progress sessions)
// and duration from start to end for ended sessions
func (s *Session) CurrentDuration() time.Duration {
	if s.endTime != nil {
		return s.endTime.Sub(s.StartTime)
	}
	return time.Since(s.StartTime)
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

	fmt.Printf("Session will end automatically in %s (at %s)\n", duration, endTime.Format("15:04:05"))

	return time.After(duration), nil
}

func (s *Session) Stop() {
	now := time.Now()
	s.endTime = &now
}

func (s *Session) EndTime() *time.Time {
	return s.endTime
}
