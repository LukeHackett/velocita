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

// TestProviderLockFileConsistency validates that .terraform.lock.hcl exists,
// is properly formatted, and contains the expected hashicorp/aws provider entry.
//
// **Validates: Requirements 3.4**
//
// This property test ensures that the provider lock file is present in the
// repository and records the hashicorp/aws provider, guaranteeing that provider
// versions and checksums are pinned for reproducible infrastructure deployments
// across all environments. The test also validates the lock file format to ensure
// it contains the required structure (provider blocks with version and hashes).
func TestProviderLockFileConsistency(t *testing.T) {
	// Find the repository root (searches for .git, falling back to Makefile + terraform/)
	repoRoot, err := findRepoRoot()
	require.NoError(t, err, "failed to locate repository root")

	// Verify .terraform.lock.hcl exists
	lockFilePath := filepath.Join(repoRoot, ".terraform.lock.hcl")
	_, err = os.Stat(lockFilePath)
	assert.NoError(t, err, ".terraform.lock.hcl file does not exist at %s", lockFilePath)

	if err != nil {
		t.Skip("Skipping lock file content check because lock file does not exist")
	}

	// Verify the lock file contains hashicorp/aws provider entry
	hasAWSProvider, err := hasProviderEntry(lockFilePath, "hashicorp/aws")
	require.NoError(t, err, "failed to read .terraform.lock.hcl")
	assert.True(t, hasAWSProvider, ".terraform.lock.hcl does not contain hashicorp/aws provider entry")

	// Verify the lock file has proper format and content
	isFormatValid, hasVersionConstraint, hasHashes, err := validateLockFileFormat(lockFilePath)
	require.NoError(t, err, "failed to validate .terraform.lock.hcl format")
	assert.True(t, isFormatValid, ".terraform.lock.hcl is not properly formatted (expected HCL provider blocks)")
	assert.True(t, hasVersionConstraint, ".terraform.lock.hcl is missing version information")
	assert.True(t, hasHashes, ".terraform.lock.hcl is missing provider checksums/hashes")
}

// hasProviderEntry checks if a provider entry exists in the lock file
// by searching for the provider identifier in the expected format.
func hasProviderEntry(lockFilePath string, provider string) (bool, error) {
	file, err := os.Open(lockFilePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Look for provider blocks in the format:
		// provider "registry.terraform.io/hashicorp/aws" {
		// or simpler format:
		// provider "hashicorp/aws" {
		if strings.Contains(line, provider) && strings.Contains(line, "provider") && strings.Contains(line, "{") {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}

// validateLockFileFormat checks if the lock file has proper HCL structure
// and contains required provider metadata (version constraints and hashes).
func validateLockFileFormat(lockFilePath string) (bool, bool, bool, error) {
	file, err := os.Open(lockFilePath)
	if err != nil {
		return false, false, false, err
	}
	defer file.Close()

	var (
		foundProviderBlock     = false
		foundVersionConstraint = false
		foundHashes            = false
	)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Check for provider blocks
		if strings.Contains(line, "provider") && strings.Contains(line, "{") {
			foundProviderBlock = true
		}

		// Check for version constraint (version or constraints field)
		if (strings.Contains(line, "version") || strings.Contains(line, "constraints")) && strings.Contains(line, "=") {
			foundVersionConstraint = true
		}

		// Check for hashes (either hashes field or h1:, h256:, etc.)
		if (strings.Contains(line, "hashes") || strings.Contains(line, "h1:") || strings.Contains(line, "h256:")) && strings.Contains(line, "[") {
			foundHashes = true
		}
	}

	if err := scanner.Err(); err != nil {
		return false, false, false, err
	}

	return foundProviderBlock, foundVersionConstraint, foundHashes, nil
}
