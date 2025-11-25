package pkg

import "fmt"

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
