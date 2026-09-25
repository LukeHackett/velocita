package test

import (
	"os"
	"path/filepath"
)

// findRepoRoot locates the repository root by searching for .git directory
// or by checking for key repository files.
func findRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Search upward from current directory
	for {
		// Check for .git directory (most reliable indicator of repo root)
		if _, err := os.Stat(filepath.Join(cwd, ".git")); err == nil {
			return cwd, nil
		}

		// Check for Makefile (project-specific)
		if _, err := os.Stat(filepath.Join(cwd, "Makefile")); err == nil {
			if _, err := os.Stat(filepath.Join(cwd, "terraform")); err == nil {
				return cwd, nil
			}
		}

		parent := filepath.Dir(cwd)
		if parent == cwd {
			// Reached filesystem root without finding repo root
			break
		}
		cwd = parent
	}

	// If not found by walking up, try current working directory
	return ".", nil
}
