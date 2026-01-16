package sheets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCredentialsLoading(t *testing.T) {
	// Create a temporary directory for test scripts
	tempDir := t.TempDir()

	// Create a mock credentials JSON
	mockCreds := `{
  "type": "service_account",
  "project_id": "test-project",
  "private_key_id": "test-key-id",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7W\n-----END PRIVATE KEY-----\n",
  "client_email": "test@test-project.iam.gserviceaccount.com",
  "client_id": "123456789",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/test%40test-project.iam.gserviceaccount.com"
}`

	t.Run("credentials from helper script", func(t *testing.T) {
		// Create a test helper script
		helperPath := filepath.Join(tempDir, "creds-helper.sh")
		helperContent := `#!/bin/bash
cat << 'EOF'
` + mockCreds + `
EOF`

		err := os.WriteFile(helperPath, []byte(helperContent), 0755)
		if err != nil {
			t.Fatalf("failed to create test helper: %v", err)
		}

		// Test that ExecuteCredentialsHelper works
		credentials, err := ExecuteCredentialsHelper(helperPath)
		if err != nil {
			t.Fatalf("failed to execute credentials helper: %v", err)
		}

		if len(credentials) == 0 {
			t.Error("expected credentials, got empty result")
		}

		// Verify it contains expected fields
		credsStr := string(credentials)
		if len(credsStr) < 100 {
			t.Errorf("credentials seem too short: %d bytes", len(credentials))
		}
	})

}
