package types

import (
	"fmt"
)

// CraftieError represents a base error type for the application
type CraftieError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Cause   error  `json:"cause,omitempty"`
}

func (e *CraftieError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *CraftieError) Unwrap() error {
	return e.Cause
}

// Error codes
const (
	ErrCodeValidation    = "VALIDATION_ERROR"
	ErrCodeDatabase      = "DATABASE_ERROR"
	ErrCodeNetwork       = "NETWORK_ERROR"
	ErrCodeAuth          = "AUTH_ERROR"
	ErrCodeConfig        = "CONFIG_ERROR"
	ErrCodeSession       = "SESSION_ERROR"
	ErrCodeDaemon        = "DAEMON_ERROR"
	ErrCodeNotification  = "NOTIFICATION_ERROR"
	ErrCodeSync          = "SYNC_ERROR"
	ErrCodeFileSystem    = "FILESYSTEM_ERROR"
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeAlreadyExists = "ALREADY_EXISTS"
	ErrCodePermission    = "PERMISSION_ERROR"
	ErrCodeTimeout       = "TIMEOUT_ERROR"
)

// ValidationError represents a configuration or input validation error
type ValidationError struct {
	*CraftieError
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		CraftieError: &CraftieError{
			Code:    ErrCodeValidation,
			Message: message,
		},
	}
}

// DatabaseError represents a database operation error
type DatabaseError struct {
	*CraftieError
}

func NewDatabaseError(message string) *DatabaseError {
	return &DatabaseError{
		CraftieError: &CraftieError{
			Code:    ErrCodeDatabase,
			Message: message,
		},
	}
}

func NewDatabaseErrorWithCause(message string, cause error) *DatabaseError {
	return &DatabaseError{
		CraftieError: &CraftieError{
			Code:    ErrCodeDatabase,
			Message: message,
			Cause:   cause,
		},
	}
}

// SessionError represents a session management error
type SessionError struct {
	*CraftieError
}

func NewSessionError(message string) *SessionError {
	return &SessionError{
		CraftieError: &CraftieError{
			Code:    ErrCodeSession,
			Message: message,
		},
	}
}

// DaemonError represents a daemon operation error
type DaemonError struct {
	*CraftieError
}

func NewDaemonError(message string) *DaemonError {
	return &DaemonError{
		CraftieError: &CraftieError{
			Code:    ErrCodeDaemon,
			Message: message,
		},
	}
}

// NotFoundError represents a resource not found error
type NotFoundError struct {
	*CraftieError
	Resource string `json:"resource"`
	ID       string `json:"id"`
}

func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{
		CraftieError: &CraftieError{
			Code:    ErrCodeNotFound,
			Message: fmt.Sprintf("%s with ID '%s' not found", resource, id),
		},
		Resource: resource,
		ID:       id,
	}
}

// Common error instances
var (
	ErrNoActiveSession      = NewSessionError("no active session found")
	ErrSessionAlreadyActive = NewSessionError("a session is already active")
	ErrDaemonNotRunning     = NewDaemonError("daemon is not running")
	ErrDaemonAlreadyRunning = NewDaemonError("daemon is already running")
)
