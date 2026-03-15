package session

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vlad/craftie/internal/notify"
)

// Recorder handles persisting session data to a storage backend
type Recorder interface {
	Record(*Session) error
}

type Session struct {
	StartTime   time.Time
	endTime     *time.Time
	ProjectName string
	Task        string
	Notes       string

	notifyEnabled      bool
	notifySoundEnabled bool
}

// SetNotifications configures notification settings for this session
func (s *Session) SetNotifications(enabled, soundEnabled bool) {
	s.notifyEnabled = enabled
	s.notifySoundEnabled = soundEnabled
}

// Run starts the session and blocks until it ends (interrupt or timer)
func (s *Session) Run(endTimerCh <-chan time.Time, syncInterval time.Duration, recorders []Recorder) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	syncChan := time.Tick(syncInterval)

	if err := s.save(recorders); err != nil {
		return fmt.Errorf("initial save failed: %w", err)
	}

loop:
	for {
		select {
		case <-sigChan:
			fmt.Println("Session interrupted")
			break loop
		case <-endTimerCh:
			fmt.Println("Session time reached!")
			s.Stop()
			break loop
		case <-syncChan:
			fmt.Printf("Syncing session (duration: %s)\n", s.FormattedDuration())
			if err := s.save(recorders); err != nil {
				fmt.Printf("Warning: sync failed: %v\n", err)
			}
		}
	}

	if err := s.save(recorders); err != nil {
		return fmt.Errorf("final save failed: %w", err)
	}

	// Notify on session end (timer or interrupt)
	if s.notifyEnabled {
		n, err := notify.OpenBusNotifySession(notify.OpenSessionParams{SoundEnabled: s.notifySoundEnabled})
		if err != nil {
			return err
		}
		defer n.Close()

		_, err = n.Send("Session Complete", fmt.Sprintf("Your craftie session for %s has ended", s.ProjectName), []string{})
		return err
	}

	return nil
}

func (s *Session) save(recorders []Recorder) error {
	for _, r := range recorders {
		if err := r.Record(s); err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) CurrentDuration() time.Duration {
	if s.endTime != nil {
		return s.endTime.Sub(s.StartTime)
	}
	return time.Since(s.StartTime)
}

func (s *Session) FormattedDuration() string {
	return time.Time{}.Add(s.CurrentDuration()).Format(time.TimeOnly)
}

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

func (s *Session) IsComplete() bool {
	return s.endTime != nil
}
