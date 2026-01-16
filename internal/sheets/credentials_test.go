package sheets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExecuteCredentialsHelper(t *testing.T) {
	// Create a temporary directory for test scripts
	tempDir := t.TempDir()

	t.Run("valid helper script", func(t *testing.T) {
		// Create a test helper script that outputs valid JSON
		helperPath := filepath.Join(tempDir, "valid-helper.sh")
		helperContent := `#!/bin/bash
echo '{"type": "service_account", "project_id": "test-project"}'`

		err := os.WriteFile(helperPath, []byte(helperContent), 0755)
		if err != nil {
			t.Fatalf("failed to create test helper: %v", err)
		}

		// Execute the helper
		result, err := ExecuteCredentialsHelper(helperPath)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		expected := `{"type": "service_account", "project_id": "test-project"}`
		if string(result) != expected {
			t.Errorf("expected %q, got %q", expected, string(result))
		}
	})

	t.Run("non-existent helper script", func(t *testing.T) {
		helperPath := filepath.Join(tempDir, "does-not-exist.sh")

		_, err := ExecuteCredentialsHelper(helperPath)
		if err == nil {
			t.Fatal("expected error for non-existent script, got nil")
		}
	})

	t.Run("non-executable helper script", func(t *testing.T) {
		// Create a script without execute permissions
		helperPath := filepath.Join(tempDir, "non-executable.sh")
		helperContent := `#!/bin/bash
echo '{"test": "data"}'`

		err := os.WriteFile(helperPath, []byte(helperContent), 0644) // No execute permission
		if err != nil {
			t.Fatalf("failed to create test helper: %v", err)
		}

		_, err = ExecuteCredentialsHelper(helperPath)
		if err == nil {
			t.Fatal("expected error for non-executable script, got nil")
		}
	})

	t.Run("helper script that fails", func(t *testing.T) {
		// Create a script that exits with non-zero status
		helperPath := filepath.Join(tempDir, "failing-helper.sh")
		helperContent := `#!/bin/bash
echo "Error: Failed to retrieve credentials" >&2
exit 1`

		err := os.WriteFile(helperPath, []byte(helperContent), 0755)
		if err != nil {
			t.Fatalf("failed to create test helper: %v", err)
		}

		_, err = ExecuteCredentialsHelper(helperPath)
		if err == nil {
			t.Fatal("expected error for failing script, got nil")
		}
	})

}
