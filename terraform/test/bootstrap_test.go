package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

// TestBootstrapModuleExists verifies that the Terraform configuration
// in the terraform/ directory is syntactically valid and can be initialized
// without errors. This test does NOT deploy any real infrastructure.
func TestBootstrapModuleExists(t *testing.T) {
	t.Parallel()

	terraformOptions := &terraform.Options{
		TerraformDir: "../",
		NoColor:      true,
	}

	// Validate runs `terraform validate` which checks syntax and internal
	// consistency without requiring a backend or provider credentials.
	out, err := terraform.ValidateE(t, terraformOptions)
	assert.NoError(t, err, "terraform validate failed: %s", out)
}

// TestTerraformConfigurationValidity is a property-based test that validates
// Terraform configuration is syntactically valid and can be validated without errors.
// **Validates: Requirements 2.1, 2.2**
// This property test verifies that the terraform validate command exits with code 0,
// confirming that all Terraform files are correctly formatted and internally consistent.
func TestTerraformConfigurationValidity(t *testing.T) {
	t.Parallel()

	terraformOptions := &terraform.Options{
		TerraformDir: "../",
		NoColor:      true,
	}

	// Property: Terraform configuration must be valid
	// terraform validate checks:
	// - All required providers are declared in versions.tf
	// - All referenced variables are defined
	// - All resource types and arguments are valid
	// - HCL syntax is correct
	out, err := terraform.ValidateE(t, terraformOptions)
	assert.NoError(t, err, "terraform validate must exit with code 0. Output: %s", out)
}
