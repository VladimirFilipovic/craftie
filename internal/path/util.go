package path

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vlad/craftie/pkg/types"
)

func ExpandPathWithHome(path string) (string, error) {
	if path == "" {
		return "", types.NewValidationError("path cannot be empty")
	}

	if !strings.HasPrefix(path, "~/") {
		return path, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", types.NewValidationError(fmt.Sprintf("failed to get home directory: %v", err))
	}

	expandedPath := filepath.Join(homeDir, path[2:])
	return expandedPath, nil
}
