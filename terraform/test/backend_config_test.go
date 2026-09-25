package test

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBackendConfigCompleteness validates that each backend.tfvars file contains
// all required keys and no credentials.
//
// **Validates: Requirements 4.2**
//
// This property test ensures that:
// 1. Each backend.tfvars file for every environment contains the required keys:
//    - bucket
//    - dynamodb_table
//    - region
//    - key
// 2. No credentials (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN)
//    are hardcoded in any backend.tfvars file.
//
// This validates that backend configuration is properly externalized and secure.
func TestBackendConfigCompleteness(t *testing.T) {
	// Find the repository root
	repoRoot, err := findRepoRoot()
	require.NoError(t, err, "failed to locate repository root")

	// Test data: environments and their expected backend.tfvars paths
	testCases := []struct {
		name        string
		backendPath string
	}{
		{
			name:        "dev environment backend config",
			backendPath: filepath.Join(repoRoot, "terraform", "environments", "dev", "backend.tfvars"),
		},
		{
			name:        "prod environment backend config",
			backendPath: filepath.Join(repoRoot, "terraform", "environments", "prod", "backend.tfvars"),
		},
	}

	requiredKeys := []string{"bucket", "dynamodb_table", "region", "key"}
	forbiddenPatterns := []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN"}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// File must exist
			_, err := os.Stat(tc.backendPath)
			require.NoError(t, err, "backend.tfvars file not found: %s", tc.backendPath)

			// Parse the backend config file
			foundKeys := make(map[string]bool)
			err = parseBackendConfig(tc.backendPath, foundKeys)
			require.NoError(t, err, "failed to parse backend.tfvars: %s", tc.backendPath)

			// Verify all required keys are present
			for _, key := range requiredKeys {
				assert.True(t, foundKeys[key],
					"missing required key '%s' in %s", key, tc.backendPath)
			}

			// Verify no forbidden patterns (credentials) are present
			err = checkForbiddenPatterns(tc.backendPath, forbiddenPatterns)
			assert.NoError(t, err,
				"forbidden credential pattern found in %s", tc.backendPath)
		})
	}
}

// parseBackendConfig reads a .tfvars file and extracts key assignments
func parseBackendConfig(path string, foundKeys map[string]bool) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Look for key = value patterns
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			foundKeys[key] = true
		}
	}

	return scanner.Err()
}

// checkForbiddenPatterns scans a file for credential-like patterns
func checkForbiddenPatterns(path string, patterns []string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip comments
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		// Check for forbidden patterns
		for _, pattern := range patterns {
			if strings.Contains(line, pattern) {
				return NewCredentialFoundError(pattern, lineNum)
			}
		}
	}

	return scanner.Err()
}

// CredentialFoundError indicates a credential was found in the backend config
type CredentialFoundError struct {
	Pattern string
	LineNum int
}

// NewCredentialFoundError creates a new credential error
func NewCredentialFoundError(pattern string, lineNum int) error {
	return &CredentialFoundError{Pattern: pattern, LineNum: lineNum}
}

// Error implements the error interface
func (e *CredentialFoundError) Error() string {
	return "credential pattern found in backend config: " + e.Pattern
}
