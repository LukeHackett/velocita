.PHONY: help init plan apply fmt lint trivy test bootstrap check-deps clean

TERRAFORM_DIR := terraform
TEST_DIR      := $(TERRAFORM_DIR)/test

# Default empty ENV variable - must be set by user for environment-scoped operations
ENV ?=

# ============================================================================
# Environment Guard
# ============================================================================

check-env:
ifndef ENV
	@echo "ERROR: ENV is not set. Usage: make <target> ENV=dev"
	@exit 1
endif

# ============================================================================
# Core Terraform Targets (Task 5.1)
# ============================================================================

init: check-env ## Initialize Terraform for the specified environment
	terraform -chdir=$(TERRAFORM_DIR) init \
	  -backend-config=environments/$(ENV)/backend.tfvars

plan: check-env ## Plan Terraform changes for the specified environment
	terraform -chdir=$(TERRAFORM_DIR) plan \
	  -var-file=environments/$(ENV)/terraform.tfvars \
	  -out=../tfplan-$(ENV)

apply: check-env ## Apply Terraform changes using the plan file
	@if [ ! -f tfplan-$(ENV) ]; then \
		echo "ERROR: Plan file tfplan-$(ENV) not found. Run 'make plan ENV=$(ENV)' first."; \
		exit 1; \
	fi
	terraform -chdir=$(TERRAFORM_DIR) apply ../tfplan-$(ENV)

fmt: ## Format all Terraform files
	terraform -chdir=$(TERRAFORM_DIR) fmt -recursive

lint: ## Lint all Terraform files
	tflint --chdir=$(TERRAFORM_DIR) --recursive

trivy: ## Scan Terraform files for security vulnerabilities
	trivy config $(TERRAFORM_DIR) --severity=HIGH,CRITICAL --ignorefile=.trivyignore.yaml

test: ## Run Terratest unit tests
	cd $(TEST_DIR) && go test -v -timeout 30m ./...

# ============================================================================
# Bootstrap and Helper Targets (Task 5.2)
# ============================================================================

bootstrap: check-env ## Deploy CloudFormation stack for backend infrastructure
	aws cloudformation deploy \
	  --stack-name velocita-$(ENV)-bootstrap \
	  --template-file bootstrap/cloudformation.yml \
	  --parameter-overrides Environment=$(ENV) ProjectName=velocita \
	  --capabilities CAPABILITY_NAMED_IAM

check-deps: ## Verify all required dependencies are installed and versions are correct
	@echo "Checking dependencies..."
	@echo ""
	
	@# Check for terraform binary
	@if command -v terraform >/dev/null 2>&1; then \
		echo "✓ terraform found"; \
	else \
		echo "✗ terraform NOT found"; \
		exit 1; \
	fi
	
	@# Check for aws CLI
	@if command -v aws >/dev/null 2>&1; then \
		echo "✓ aws found"; \
	else \
		echo "✗ aws NOT found"; \
		exit 1; \
	fi
	
	@# Check for tflint
	@if command -v tflint >/dev/null 2>&1; then \
		echo "✓ tflint found"; \
	else \
		echo "✗ tflint NOT found"; \
		exit 1; \
	fi
	
	@# Check for trivy
	@if command -v trivy >/dev/null 2>&1; then \
		echo "✓ trivy found"; \
	else \
		echo "✗ trivy NOT found"; \
		exit 1; \
	fi
	
	@# Check for go
	@if command -v go >/dev/null 2>&1; then \
		echo "✓ go found"; \
	else \
		echo "✗ go NOT found"; \
		exit 1; \
	fi
	
	@# Check for jq
	@if command -v jq >/dev/null 2>&1; then \
		echo "✓ jq found"; \
	else \
		echo "✗ jq NOT found"; \
		exit 1; \
	fi
	
	@echo ""
	@echo "All dependencies verified!"

clean: ## Clean auto-generated files and reset project to clean state
	@echo "Cleaning auto-generated files..."
	
	@# Remove Terraform working directories
	@rm -rf $(TERRAFORM_DIR)/.terraform
	@echo "  ✓ Removed .terraform directories"
	
	@# Remove Terraform plan files
	@rm -f tfplan-*
	@echo "  ✓ Removed plan files (tfplan-*)"
	
	@# Remove Terraform lock files (but keep checked-in lock files)
	@find $(TERRAFORM_DIR) -name ".terraform.lock.hcl" -type f -delete
	@echo "  ✓ Removed .terraform.lock.hcl"
	
	@# Remove Go build and test artifacts in test directory
	@cd $(TEST_DIR) && go clean
	@rm -f $(TEST_DIR)/*.out $(TEST_DIR)/*.test $(TEST_DIR)/*.cover
	@echo "  ✓ Cleaned Go build artifacts"
	
	@# Remove Go module cache
	@rm -f $(TEST_DIR)/go.work.sum
	@echo "  ✓ Cleaned Go module cache"
	
	@# Remove test coverage files
	@rm -f $(TEST_DIR)/coverage.out $(TEST_DIR)/coverage.html
	@echo "  ✓ Removed test coverage files"
	
	@echo ""
	@echo "Project cleaned successfully!"
	@echo "To reinitialize: make init ENV=dev"

help: ## Display this help message
	@echo "velocita - Terraform Project Bootstrap"
	@echo ""
	@echo "Usage: make [target] [ENV=dev|prod]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -v '^check-env' | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}' | sort
	@echo ""
	@echo "Examples:"
	@echo "  make check-deps             # Verify dependencies"
	@echo "  make bootstrap ENV=dev      # Deploy backend infrastructure"
	@echo "  make init ENV=dev           # Initialize Terraform"
	@echo "  make plan ENV=dev           # Plan infrastructure changes"
	@echo "  make apply ENV=dev          # Apply infrastructure changes"
	@echo "  make fmt                    # Format Terraform files"
	@echo "  make lint                   # Lint Terraform files"
	@echo "  make trivy                  # Scan for security vulnerabilities"
	@echo "  make test                   # Run tests"
	@echo "  make clean                  # Clean all auto-generated files"
