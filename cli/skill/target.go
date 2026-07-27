package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetTargetDirectory resolves the installation directory based on scope, agent, and skill name.
func GetTargetDirectory(scope string, agent string, name string) (string, error) {
	baseDir := ""

	switch scope {
	case "project", "local":
		baseDir = ".agents/skills"
	case "user", "global":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not determine home directory: %w", err)
		}

		switch agent {
		case "cursor":
			baseDir = filepath.Join(home, ".cursor", "skills")
		case "copilot":
			baseDir = filepath.Join(home, ".github-copilot", "skills")
		default:
			baseDir = filepath.Join(home, ".agents", "skills")
		}
	default:
		return "", fmt.Errorf("unknown scope: %s (must be user or project)", scope)
	}

	if name != "" {
		// Prevent path traversal
		cleanName := filepath.Clean(name)
		if strings.Contains(cleanName, "..") || strings.Contains(cleanName, string(os.PathSeparator)) || cleanName == "/" {
			return "", fmt.Errorf("invalid skill name: %s", name)
		}
		return filepath.Join(baseDir, cleanName), nil
	}
	return baseDir, nil
}
