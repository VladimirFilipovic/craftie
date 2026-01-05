package session

import (
	"time"
)

type Session struct {
	StartTime   time.Time
	ProjectName string
	Notes       string
}

func (s *Session) GetDurationSec() uint64 {
	return uint64(time.Since(s.StartTime).Seconds())
}
