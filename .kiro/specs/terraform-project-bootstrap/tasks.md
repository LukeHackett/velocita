# Implementation Plan: Terraform Project Bootstrap

## Overview

This implementation plan converts the design into discrete, incremental coding tasks. The bootstrap establishes a fully configured Terraform project foundation with version pinning, remote state management, a Makefile for developer workflows (including local `fmt`/`lint`/`trivy`/`clean` targets), GitHub Actions CI/CD with a multi-job pre-merge architecture (lint, security, plan, comment jobs running in parallel with aggregated PR comments), AWS OIDC authentication via repository variables, artifact storage for Terraform plans, composite actions for reusability, a CloudFormation bootstrap template, and a Terratest unit testing framework.

The tasks are ordered to build incrementally: first directory structure and basic configuration files, then version pinning and provider setup, followed by backend configuration, Makefile targets, composite GitHub Actions, workflows (pre-merge with multi-job architecture first, then post-merge), CloudFormation bootstrap template, and finally the Terratest scaffold. Two checkpoints ensure progress validation.

## Tasks

- [x] 1. Set up directory structure and foundational configuration files
  - [x] 1.1 Create directory structure and tracked placeholder files
    - Create `terraform/modules/.gitkeep`
    - Create `terraform/environments/dev/.gitkeep`
    - Create `terraform/environments/prod/.gitkeep`
    - Create `bootstrap/` directory (no placeholder needed)
    - _Requirements: 1.1, 1.2, 1.3_
  
  - [x] 1.2 Create root-level `.gitattributes` file
    - Set `*.tf` and `*.tfvars` to `text eol=lf`
    - Set `*.hcl` to `text eol=lf`
    - Mark `.terraform.lock.hcl` as `linguist-generated=true`
    - Mark `tfplan-*` files as binary
    - _Requirements: 1.5_
  
  - [x] 1.3 Create root-level `.editorconfig` file
    - Set default indent, line endings, charset, trailing whitespace, and final newline rules
    - Override `indent_size = 2` for `*.tf`, `*.tfvars`, `*.yml`, `*.yaml`, `*.json` files
    - _Requirements: 1.6_
  
  - [x] 1.4 Create root-level `.gitignore` file
    - Ignore `.terraform/` directories
    - Ignore `tfplan-*` plan output files
    - Ignore `*.auto.tfvars` files
    - Ignore `crash.log` and `.terraform.tfstate*` files
    - _Requirements: 1.7_
  
  - [x] 1.5 Create root-level `README.md` file
    - Document directory structure overview
    - Explain purpose of `terraform/modules/` and `terraform/environments/`
    - Document contributor onboarding steps: tool installation, `make check-deps`, `make bootstrap ENV=dev`, `make init ENV=dev`
    - Document how to run the test suite (`make test`)
    - _Requirements: 1.4_
  
  - [x] 1.6 Create root-level `LICENSE` file
    - Add MIT License text with placeholder for copyright year and holder
    - _Requirements: 1.10_

- [x] 2. Set up remote state backend configuration
  - [x] 1.1 Create `terraform/backend.tf` file
    - Define S3 backend with empty configuration body (values injected via `-backend-config`)
    - Include documentation comment explaining per-environment backend.tfvars usage
    - _Requirements: 4.1, 4.4_
  
  - [x] 1.2 Create `terraform/environments/dev/backend.tfvars` file
    - Specify `bucket = "velocita-dev-terraform-state"`
    - Specify `dynamodb_table = "velocita-dev-terraform-locks"`
    - Specify `region = "eu-west-1"`
    - Specify `key = "dev/terraform.tfstate"`
    - _Requirements: 4.2, 1.8_
  
  - [x] 1.3 Create `terraform/environments/prod/backend.tfvars` file
    - Specify `bucket = "velocita-prod-terraform-state"`
    - Specify `dynamodb_table = "velocita-prod-terraform-locks"`
    - Specify `region = "eu-west-1"`
    - Specify `key = "prod/terraform.tfstate"`
    - _Requirements: 4.2, 1.8_
  
  - [x] 1.4 Create `terraform/environments/dev/terraform.tfvars` file
    - Create as empty placeholder file with comment
    - _Requirements: 1.9_
  
  - [x] 1.5 Create `terraform/environments/prod/terraform.tfvars` file
    - Create as empty placeholder file with comment
    - _Requirements: 1.9_
  
  - [x] 1.6 Write property test for backend config completeness
    - **Property 3: Backend Config Completeness**
    - **Validates: Requirements 4.2**
    - Test that each `backend.tfvars` file contains all required keys (bucket, dynamodb_table, region, key) and no credentials

- [x] 2. Create Makefile with developer targets
  - [x] 1.1 Create `Makefile` at repository root
    - Implement `init` target: `terraform init -backend-config=environments/<ENV>/backend.tfvars` with ENV guard
    - Implement `plan` target: `terraform plan -var-file=environments/<ENV>/terraform.tfvars -out=tfplan-<ENV>` with ENV guard
    - Implement `apply` target: apply the plan file with guard for plan file existence
    - Implement `fmt` target: `terraform fmt -recursive`
    - Implement `lint` target: `tflint --recursive`
    - Implement `test` target: `go test -v -timeout 30m ./...` in `terraform/test/` directory
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7, 5.8, 5.9_
  
  - [x] 1.2 Add bootstrap and helper targets to Makefile
    - Implement `bootstrap` target: `aws cloudformation deploy` for CloudFormation stack creation with ENV guard
    - Implement `check-deps` target: verify presence of terraform, aws, tflint, trivy, go, jq
    - Implement `clean` target: remove `.terraform/` directories, `tfplan-*` files, `.terraform.lock.hcl`, and Go build/test artifacts
    - Implement `help` target: display all available targets with descriptions
    - _Requirements: 5.10, 5.11, 5.12, 5.13, 5.14, 5.15, 1.8 (bootstrap)_

- [x] 2. Create GitHub Actions composite actions
  - [x] 1.1 Create `.github/actions/aws-oidc-auth/action.yml` composite action
    - Define inputs: `role-arn` (required), `aws-region` (optional, default: eu-west-1)
    - Use `aws-actions/configure-aws-credentials@v4` step with OIDC role assumption
    - Set `role-session-name` to include `github.run_id` for CloudTrail auditability
    - _Requirements: 6.4, 6.20, 6.21, 6.22, 6.23_
  
  - [x] 1.2 Create `.github/actions/setup-terraform/action.yml` composite action
    - Define inputs: `environment` (required, default: dev), `tool-versions-file` (optional, default: `.tool-versions`)
    - Add a step that parses the pinned `terraform` entry out of `.tool-versions` and exposes it as a step output, since `hashicorp/setup-terraform@v3` has no file-based version input
    - Define outputs: `terraform-version` (from the version-parsing step)
    - Use `hashicorp/setup-terraform@v3` action with the parsed `terraform_version` and `terraform_wrapper: true`
    - Add `terraform init -backend-config=environments/<environment>/backend.tfvars` step
    - _Requirements: 6.4, 7.18_

- [x] 2. Create GitHub Actions pre-merge workflow with multi-job architecture
  - [x] 1.1 Create `.github/workflows/pre-merge.yml` workflow file
    - Define trigger: `pull_request` events targeting `main` branch
    - Set permissions: `contents: read`, `id-token: write`, `pull-requests: write`
    - Define environment variable `ENV: dev`
    - _Requirements: 6.1, 7.7_
  
  - [x] 1.2 Implement `lint` job in pre-merge workflow
    - Run on `ubuntu-latest`
    - Checkout code
    - Call `aws-oidc-auth` composite action with `vars.AWS_PLAN_ROLE_ARN` (repository variable, not secret)
    - Call `setup-terraform` composite action
    - Run `terraform fmt -check -recursive` (id: fmt)
    - Run `tflint --recursive` (id: tflint)
    - Job must exit non-zero if either check fails
    - _Requirements: 6.8, 7.18, 7.21_
  
  - [x] 1.3 Implement `security` job in pre-merge workflow
    - Run on `ubuntu-latest`
    - Checkout code
    - Run `aquasecurity/trivy-action@v0.30.0` with config scan, terraform/ directory, HIGH,CRITICAL severity, exit-code 1
    - Job may fail if findings detected
    - _Requirements: 6.9, 7.18_
  
  - [x] 1.4 Implement `plan` job in pre-merge workflow
    - Run on `ubuntu-latest`
    - Checkout code
    - Call `aws-oidc-auth` composite action with `vars.AWS_PLAN_ROLE_ARN` (repository variable, not secret)
    - Call `setup-terraform` composite action
    - Run `terraform plan -var-file=environments/<ENV>/terraform.tfvars -out=../tfplan-<ENV> -no-color` (id: plan)
    - Export plan output as job output `plan-summary` for comment job to consume
    - Upload plan artifact with 1-day retention using `actions/upload-artifact@v4`
    - _Requirements: 6.10, 7.18, 7.21_
  
  - [x] 1.5 Implement `comment` job in pre-merge workflow
    - Run on `ubuntu-latest`
    - Depend on: `lint`, `security`, `plan` jobs via `needs: [lint, security, plan]`
    - Use `if: always()` to run even if upstream jobs fail
    - Checkout code
    - Download plan artifact from `plan` job using `actions/download-artifact@v4`
    - Use `GetTerminus/terraform-pr-commenter@v3` to post aggregated PR comment
    - Post comment includes results from fmt, tflint, trivy, and terraform plan
    - Use `needs.*.result` to aggregate job statuses
    - Exit non-zero if any upstream job failed (using `if:` condition at end of job)
    - _Requirements: 6.11, 7.12, 7.13, 7.14, 7.15_
  
  - [x] 1.6 Implement job parallelism in pre-merge workflow
    - Lint, security, and plan jobs have no inter-job dependencies
    - Lint, security, and plan jobs run in parallel (no `needs:` clause)
    - Comment job uses `needs: [lint, security, plan]` to ensure parallel jobs complete first
    - _Requirements: 6.11_
  
  - [x] 1.7 Pin GitHub Actions versions in pre-merge workflow
    - Pin all actions to major version tags (`@v4`, `@v3`, `@v0.30.0`, not commit SHAs or branches)
    - _Requirements: 6.18_

- [x] 2. Create GitHub Actions post-merge workflow
  - [x] 1.1 Create `.github/workflows/post-merge.yml` workflow file
    - Define trigger: `push` events to `main` branch
    - Set permissions: `contents: read`, `id-token: write`
    - Define environment variable `ENV: dev`
    - _Requirements: 6.2, 7.16_
  
  - [x] 1.2 Implement sequential verification steps in post-merge workflow
    - Single job: `verify`
    - Run on `ubuntu-latest`
    - Checkout code
    - Call `aws-oidc-auth` composite action with `vars.AWS_INIT_ROLE_ARN` (repository variable, not secret)
    - Call `setup-terraform` composite action
    - Setup Go using `actions/setup-go@v5` with go-version from `terraform/test/go.mod`
    - Run `go test -v -timeout 30m ./...` in `terraform/test/` directory
    - No `continue-on-error` on any step; halt on first failure
    - _Requirements: 6.16, 7.17, 7.22_
  
  - [x] 1.3 Pin GitHub Actions versions in post-merge workflow
    - Pin all actions to major version tags (`@v4`, `@v3`, `@v5`, not commit SHAs or branches)
    - _Requirements: 6.18_

- [x] 2. Create GitHub Dependabot configuration
  - [x] 1.1 Create `.github/dependabot.yml` file
    - Configure weekly GitHub Actions version updates
    - Set `open-pull-requests-limit: 5`
    - _Requirements: 6.19_

- [x] 2. Create CloudFormation backend bootstrap template
  - [x] 1.1 Create `bootstrap/cloudformation.yml` file
    - Define CloudFormation template version and description
    - Define parameters: `Environment` (dev, prod), `ProjectName` (default: velocita)
    - Create S3 bucket resource with deterministic name: `<ProjectName>-<Environment>-terraform-state`
    - Enable S3 versioning, AES-256 encryption, block all public access
    - Create DynamoDB table resource with name: `<ProjectName>-<Environment>-terraform-locks`
    - Use `LockID` as partition key with `String` type
    - Use `PAY_PER_REQUEST` billing mode
    - Export S3 bucket name and DynamoDB table name as stack outputs
    - _Requirements: 6.1, 9.2, 9.3, 9.4, 9.5_

- [x] 2. Set up Terratest unit testing framework
  - [x] 1.1 Create `terraform/test/` directory and Go module files
    - Create `terraform/test/go.mod` with module declaration and Terratest dependencies pinned to exact versions
    - Include github.com/gruntwork-io/terratest v0.46.16 and github.com/stretchr/testify v1.9.0
    - Create `terraform/test/go.sum` with checksums (generated by `go mod tidy`)
    - _Requirements: 6.1, 8.2, 8.3_
  
  - [x] 11.2 Create example test scaffold in `terraform/test/bootstrap_test.go`
    - Write `TestBootstrapModuleExists` test function
    - Call `terraform.ValidateE` against `../` directory
    - Assert zero exit code and no error
    - Include test documentation comment
    - _Requirements: 6.5, 8.6_
  
  - [x] 11.3 Write property test for Terraform configuration validity
    - **Property 4: Terraform Configuration Validity**
    - **Validates: Requirements 3.1, 3.2**
    - Create test that validates `terraform validate` exits 0
    - Runs as part of Terratest suite
  
  - [x] 11.4 Write property test for provider lock file consistency
    - **Property 5: Provider Lock File Consistency**
    - **Validates: Requirements 3.4**
    - Test that `.terraform.lock.hcl` exists and contains hashicorp/aws provider entry

- [x] 12. Checkpoint - Verify foundational setup
  - Ensure all files created in tasks 1–11 are present and syntactically valid
  - Run `make check-deps` and verify all required binaries are found
  - Run `make fmt`, `make lint`, and `make trivy` and verify no errors (or only expected, fixable errors)
  - Ask the user if questions arise before proceeding to AWS bootstrap

- [x] 13. Initialize Terraform and generate lock file
  - [x] 13.1 Manual step: Run `make init ENV=dev`
    - This initializes the dev environment backend
    - Generates `.terraform.lock.hcl` with provider checksums for multiple platforms
    - Commit `.terraform.lock.hcl` to version control
    - _Requirements: 3.3, 3.4_
  
  - [x] 13.2 Write property test for lock file presence and format
    - **Property 5: Provider Lock File Consistency**
    - **Validates: Requirements 3.4**
    - Test that `.terraform.lock.hcl` exists after `terraform init`

- [x] 14. Final checkpoint - Ensure all tests pass
  - Run `make test` and verify `TestBootstrapModuleExists` and other tests pass
  - Run `make fmt`, `make lint` and verify no errors
  - Verify `.terraform.lock.hcl` is committed and contains expected provider entries
  - Ask the user if questions arise before considering bootstrap complete

## Notes

- Tasks marked with `*` are optional and can be skipped for an MVP bootstrap, though recommended for comprehensive validation
- All file paths are relative to the repository root (`velocita/`)
- The `ENV` variable in the Makefile is case-sensitive and must be `dev` or `prod`
- AWS OIDC role ARNs (`vars.AWS_PLAN_ROLE_ARN`, `vars.AWS_INIT_ROLE_ARN`) are stored as GitHub **repository variables** (not secrets) because they are non-sensitive resource identifiers and are safe to log
- The multi-job pre-merge workflow runs `lint`, `security`, and `plan` jobs in parallel for minimal feedback latency, with the `comment` job aggregating all results into a single PR comment
- The plan artifact is stored with 1-day retention to support future manual apply workflows
- The post-merge workflow uses sequential steps to halt on first failure; this is intentional since there is no value in running tests if `terraform init` fails
- All GitHub Actions are pinned to major version tags to ensure reproducible behavior across workflow runs
- The bootstrap is a one-time operation; all subsequent specs (S3/Glue, Firehose/Lambda, API Gateway) depend on this foundation

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "1.2", "1.3", "1.4", "1.5", "1.6"] },
    { "id": 1, "tasks": ["2.1", "2.2", "2.3"] },
    { "id": 2, "tasks": ["3.1", "3.2", "3.3", "3.4", "3.5"] },
    { "id": 3, "tasks": ["4.1", "4.2"] },
    { "id": 4, "tasks": ["5.1", "5.2"] },
    { "id": 5, "tasks": ["6.1", "6.2"] },
    { "id": 6, "tasks": ["7.1", "7.2", "7.3", "7.4", "7.5", "7.6", "7.7"] },
    { "id": 7, "tasks": ["8.1", "8.2", "8.3"] },
    { "id": 8, "tasks": ["9.1"] },
    { "id": 9, "tasks": ["10.1"] },
    { "id": 10, "tasks": ["11.1", "11.2"] },
    { "id": 11, "tasks": ["2.4", "3.6", "11.3", "11.4"] },
    { "id": 12, "tasks": ["13.1"] },
    { "id": 13, "tasks": ["13.2"] }
  ]
}
```
