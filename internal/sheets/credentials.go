package sheets

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/zalando/go-keyring"
)

// GetCredentials fetches Google service account credentials.
// If credentialsHelper is provided, it will be used to fetch credentials,
// otherwise falls back to the system keyring.
func GetCredentials(credentialsHelper string) ([]byte, error) {
	if credentialsHelper != "" {
		credentials, err := ExecuteCredentialsHelper(credentialsHelper)
		if err != nil {
			return nil, fmt.Errorf("failed to get credentials from helper: %w", err)
		}
		return credentials, nil
	}

	// Fall back to keyring
	credsStr, err := keyring.Get("craftie", "google-sheets")
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials from keyring: %w", err)
	}
	return []byte(credsStr), nil
}

// ExecuteCredentialsHelper runs the credentials helper script and returns the output
func ExecuteCredentialsHelper(helperPath string) ([]byte, error) {
	// Check if helper script exists and is executable
	info, err := os.Stat(helperPath)
	if err != nil {
		return nil, fmt.Errorf("credentials helper not found: %w", err)
	}

	// Check if it's executable (Unix permissions)
	if info.Mode()&0111 == 0 {
		return nil, fmt.Errorf("credentials helper is not executable: %s (run: chmod +x %s)", helperPath, helperPath)
	}

	// Execute the helper script
	cmd := exec.Command(helperPath)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("credentials helper failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute credentials helper: %w", err)
	}

	// Trim whitespace from output
	return []byte(strings.TrimSpace(string(output))), nil
}
