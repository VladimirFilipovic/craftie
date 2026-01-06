package session

import (
	"fmt"
	"time"
)

type Session struct {
	StartTime   time.Time
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
