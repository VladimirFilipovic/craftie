package sheets

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

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
