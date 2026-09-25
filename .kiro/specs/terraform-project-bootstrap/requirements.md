# Requirements Document

## Introduction

This document defines the requirements for bootstrapping the `velocita` Terraform project repository. The goal is to establish a clean, modern Terraform project foundation that enforces consistent code quality, enables automated CI/CD workflows, provides remote state management, and includes a unit testing framework — all before any infrastructure-specific modules are authored. This bootstrap spec is a prerequisite for all subsequent feature specs in the `velocita` project.

## Glossary

- **Repository**: The `velocita` Git repository containing all Terraform source code.
- **Terraform_Project**: The collection of Terraform configuration files, modules, and environments managed within the Repository.
- **Module**: A reusable, self-contained unit of Terraform configuration stored under `terraform/modules/`.
- **Environment**: A deployment target (e.g., `dev`, `prod`) represented by a directory under `terraform/environments/`.
- **Backend**: Remote Terraform state storage backed by an Amazon S3 bucket and a DynamoDB table for state locking.
- **CI_Pipeline**: The GitHub Actions workflow that validates and applies Terraform changes automatically.
- **Lint**: Static analysis of Terraform code for style, correctness, and security issues.
- **Unit_Test**: An automated test executed by Terratest or an equivalent Go-based framework that validates Terraform module behavior without deploying real cloud infrastructure.
- **Composite_Action**: A reusable GitHub Actions action defined as a `composite` type under `.github/actions/`, combining multiple steps into a single callable unit.
- **OIDC_Role**: An AWS IAM role with a trust policy scoped to GitHub Actions' OIDC identity provider, used to grant temporary AWS credentials without storing long-lived secrets.

---

## Requirements

### Requirement 1: Directory Structure

**User Story:** As a developer, I want a standard Terraform directory layout, so that all contributors have a consistent place to put modules and environment configurations.

#### Acceptance Criteria

1. THE Terraform_Project SHALL contain a `terraform/modules/` directory with at least a tracked placeholder file (e.g., `.gitkeep`) so the directory is committed to the repository.
2. THE Terraform_Project SHALL contain a `terraform/environments/dev/` directory with at least a tracked placeholder file so the directory is committed to the repository.
3. THE Terraform_Project SHALL contain a `terraform/environments/prod/` directory with at least a tracked placeholder file so the directory is committed to the repository.
4. THE Repository SHALL contain a root-level `README.md` file that includes at minimum: a description of the directory structure, the purpose of `terraform/modules/` and `terraform/environments/`, and contributor onboarding steps (tool installation, first-time init, and running the test suite).
5. THE Repository SHALL contain a `.gitattributes` file that sets `*.tf` and `*.tfvars` files to `text eol=lf`, marks `.terraform.lock.hcl` as `linguist-generated=true`, and marks `tfplan-*` files as binary.
6. THE Repository SHALL contain a `.editorconfig` file that sets the following defaults for all file types: `indent_style = space`, `end_of_line = lf`, `charset = utf-8`, `trim_trailing_whitespace = true`, `insert_final_newline = true`; and overrides `indent_size = 2` for `*.tf`, `*.tfvars`, `*.yml`, `*.yaml`, and `*.json` files.
7. THE Repository SHALL contain a `.gitignore` file that ignores at minimum: `.terraform/` directories, `tfplan-*` plan output files, `*.auto.tfvars` override files, `crash.log` Terraform crash logs, and `.terraform.tfstate` local state files.
8. THE Terraform_Project SHALL contain a `terraform/environments/dev/backend.tfvars` file and a `terraform/environments/prod/backend.tfvars` file, each committed to version control and containing the non-secret backend configuration values for that environment (S3 bucket name, DynamoDB table name, AWS region, and state file key path) — no credentials or account IDs.
9. THE Terraform_Project SHALL contain a `terraform/environments/dev/terraform.tfvars` file and a `terraform/environments/prod/terraform.tfvars` file as tracked placeholder files (may be empty or contain only comments) to indicate the intended location for future environment-specific variable values.
10. THE Repository SHALL contain a `LICENSE` file at the root of the project containing the MIT License text, with the copyright year and holder left as a placeholder (e.g., `Copyright (c) [YEAR] Luke Hackett`) for the repository owner to complete.

---

### Requirement 2: Provider Version Pinning

**User Story:** As a developer, I want all Terraform provider versions pinned in a `versions.tf` file, so that provider upgrades are explicit and reproducible across all environments.

#### Acceptance Criteria

1. THE Terraform_Project SHALL contain a `terraform/versions.tf` file with a `terraform` block declaring a `required_version` constraint using `>=` or `~>` operators pinned to at minimum the minor version (e.g., `~> 1.9`).
2. THE `versions.tf` file SHALL declare all required providers in a `required_providers` block using either exact (`= X.Y.Z`) or pessimistic constraint (`~> X.Y.Z`) version operators — no open-ended `>=` constraints without an upper bound on providers.
3. IF a required provider is referenced in any `.tf` file but is not declared in the `required_providers` block of `versions.tf`, THEN `terraform init` SHALL exit with a non-zero status code and print a provider requirement error.
4. THE Repository SHALL contain a `.terraform.lock.hcl` file committed to version control, recording the exact provider versions and checksums resolved at `terraform init` time to enforce cross-environment reproducibility.

---

### Requirement 3: Remote State Backend

**User Story:** As a developer, I want Terraform state stored remotely in S3 with DynamoDB locking, so that multiple contributors can safely run Terraform without state conflicts.

#### Acceptance Criteria

1. THE Terraform_Project SHALL contain a `terraform/backend.tf` file (or per-environment `backend.tf`) configuring the `s3` backend type.
2. THE `backend.tf` file SHALL specify an S3 bucket name, a DynamoDB table name for state locking, the AWS region, and a state file key path that includes an environment-specific segment (e.g., `dev/terraform.tfstate` or `prod/terraform.tfstate`) to prevent cross-environment state collisions.
3. WHEN two concurrent `terraform apply` operations target the same state file, THE Backend SHALL prevent simultaneous writes by acquiring a DynamoDB lock and returning a lock-acquisition error to the second caller.
4. THE `backend.tf` file SHALL NOT hardcode AWS account IDs, AWS access key IDs, AWS secret access keys, or AWS session tokens; all of these values SHALL be supplied via environment variables or a `-backend-config` override file passed at `terraform init` time.
5. IF a `terraform apply` or `terraform plan` operation terminates abnormally while holding a DynamoDB state lock, THEN the lock SHALL remain in DynamoDB and be releasable by a developer running `terraform force-unlock <lock-id>` without any manual DynamoDB intervention.
6. WHEN a developer runs `make init`, THE Makefile SHALL pass the per-environment `backend.tfvars` file (e.g., `terraform/environments/<ENV>/backend.tfvars`) to `terraform init` via the `-backend-config` flag so that no backend values need to be supplied manually.

---

### Requirement 4: Makefile Developer Targets

**User Story:** As a developer, I want a Makefile with standard targets, so that I can run common Terraform operations with short, memorable commands.

#### Acceptance Criteria

1. THE Repository SHALL contain a `Makefile` at the root of the project.
2. WHEN a developer runs `make init`, THE Makefile SHALL execute `terraform init` for the environment identified by the `ENV` variable.
3. WHEN a developer runs `make plan`, THE Makefile SHALL execute `terraform plan` for the environment identified by the `ENV` variable and write the plan output to a file named `tfplan-<ENV>` in the root of the project.
4. WHEN a developer runs `make apply`, THE Makefile SHALL execute `terraform apply` using the plan file named `tfplan-<ENV>` for the environment identified by the `ENV` variable.
5. WHEN a developer runs `make fmt`, THE Makefile SHALL execute `terraform fmt -recursive` across the entire `terraform/` directory.
6. WHEN a developer runs `make lint`, THE Makefile SHALL execute `tflint --recursive` across the entire `terraform/` directory.
7. WHEN a developer runs `make trivy`, THE Makefile SHALL execute `trivy config` on the `terraform/` directory with HIGH and CRITICAL severity filtering, using the `.trivyignore.yaml` file for exceptions.
8. WHEN a developer runs `make test`, THE Makefile SHALL execute the unit test suite located under `terraform/test/`.
9. IF the `ENV` variable is not set when invoking `make init`, `make plan`, or `make apply`, THEN THE Makefile SHALL print an error message indicating the missing variable and exit with a non-zero status code.
10. IF the plan file `tfplan-<ENV>` does not exist when a developer runs `make apply`, THEN THE Makefile SHALL print an error message indicating the missing plan file and exit with a non-zero status code.
11. WHEN a developer runs `make help`, THE Makefile SHALL print a list of all available targets with a short description of each and exit with a zero status code.
12. WHEN a developer runs `make bootstrap`, THE Makefile SHALL deploy the CloudFormation stack defined in `bootstrap/cloudformation.yml` using the AWS CLI, creating or updating the S3 state bucket and DynamoDB lock table, and SHALL require the `ENV` variable to be set.
13. IF the `ENV` variable is not set when invoking `make bootstrap`, THEN THE Makefile SHALL print an error message indicating the missing variable and exit with a non-zero status code.
14. WHEN a developer runs `make check-deps`, THE Makefile SHALL verify that the following binaries are present on PATH: `terraform`, `aws`, `tflint`, `trivy`, `go`, and `jq`; and SHALL print a confirmation message for each binary found or an error message for each binary not found.
15. WHEN a developer runs `make clean`, THE Makefile SHALL remove all auto-generated files including: `.terraform/` directories, `tfplan-*` plan files, `.terraform.lock.hcl` lock files, Go build artifacts, Go module cache files, and test coverage reports, leaving the project in a clean state ready for re-initialization.
16. IF any required binary is missing, THEN `make check-deps` SHALL exit with a non-zero status code after printing all findings.

---

### Requirement 5: GitHub Actions CI Workflows

**User Story:** As a developer, I want two GitHub Actions workflows — one that runs on pull requests and one that runs on merge to main — so that infrastructure changes are validated before merge and verified after merge, with results surfaced directly in the pull request.

#### Acceptance Criteria

**Workflow Files**

1. THE Repository SHALL contain a pre-merge workflow file at `.github/workflows/pre-merge.yml` that is triggered on `pull_request` events targeting the `main` branch.
2. THE Repository SHALL contain a post-merge workflow file at `.github/workflows/post-merge.yml` that is triggered on `push` events to the `main` branch.
3. THE Repository SHALL NOT perform `terraform apply` automatically on merge to `main`; all applies SHALL be performed manually via `make apply`.

**Composite Actions**

4. THE Repository SHALL contain a composite action at `.github/actions/setup-terraform/action.yml` that installs Terraform, runs `terraform init` with the per-environment `backend.tfvars` file, and exposes Terraform step outputs for use by the PR commenter.

4a. THE `setup-terraform` composite action SHALL install the exact Terraform version pinned in the repository's `.tool-versions` file, resolving that version itself before invoking `hashicorp/setup-terraform`, since that upstream action does not support reading an asdf-style `.tool-versions` file directly.
5. THE Repository SHALL contain a composite action at `.github/actions/aws-oidc-auth/action.yml` that performs GitHub OIDC authentication and assumes the specified AWS IAM role using `aws-actions/configure-aws-credentials`.
6. THE `.github/actions/` composite actions SHALL be reusable across both the pre-merge and post-merge workflows and any future workflows.

**Pre-Merge Workflow (`pre-merge.yml`) — Multi-Job Architecture**

7. WHEN a pull request is opened or updated targeting `main`, THE pre-merge workflow SHALL define four jobs: `lint`, `security`, `plan`, and `comment`.
8. THE `lint` job SHALL run `terraform fmt -check` and `tflint` on all Terraform files; both checks must pass for the job to succeed.
9. THE `security` job SHALL run `trivy config` on all staged Terraform files; this job MAY be configured to warn (exit code 0) or fail (exit code non-zero) depending on severity thresholds.
10. THE `plan` job SHALL run `terraform init` and `terraform plan`, store the plan output as a GitHub Actions artifact, and export the plan text as a job output for the comment job to consume.
11. THE `lint`, `security`, and `plan` jobs SHALL run in parallel (no inter-job dependencies) to minimize workflow execution time.
12. THE `comment` job SHALL depend on the `lint`, `security`, and `plan` jobs; it SHALL execute only after all three jobs complete (regardless of success or failure status).
13. WHEN the `comment` job executes, THE job SHALL use `GetTerminus/terraform-pr-commenter@v3` to post an aggregated pull request comment containing the results from `terraform fmt`, `tflint`, `trivy config`, and `terraform plan`, replacing any previous comment from the same action to keep the PR timeline clean.
14. IF the `lint` or `plan` job exits with a non-zero exit code, THE `comment` job SHALL still execute and surface those failures in the PR comment.
15. WHEN any of the `lint` or `plan` jobs fails, THE overall pre-merge workflow SHALL be marked as failed and prevent merging (subject to GitHub branch protection rules).

**Post-Merge Workflow (`post-merge.yml`)**

16. WHEN a commit is pushed to `main`, THE post-merge workflow SHALL execute the following steps in sequence: AWS OIDC auth (read-only role), `terraform init`, and the Terratest unit test suite via `go test -v -timeout 30m ./...` in `terraform/test/`.
17. IF any post-merge step exits with a non-zero exit code, THEN THE post-merge workflow SHALL mark the workflow run as failed and halt subsequent steps.

**Action Version Pinning**

18. ALL GitHub Actions used in workflows and composite actions SHALL be pinned to their major version tag (e.g. `actions/checkout@v4`, `hashicorp/setup-terraform@v3`) rather than specific patch versions or commit SHAs.
19. THE `.github/dependabot.yml` file SHALL be configured to open weekly pull requests to update pinned GitHub Actions major versions when new major versions are released.

**AWS OIDC Authentication**

20. THE CI workflows SHALL authenticate to AWS using GitHub OIDC with no long-lived AWS access keys stored as GitHub secrets.
21. THE pre-merge workflow SHALL assume a read-only IAM role sufficient for `terraform plan` operations.
22. THE post-merge workflow SHALL assume a read-only IAM role sufficient for `terraform init` and test operations.
23. WHEN AWS OIDC authentication is used, THE temporary AWS credentials (STS tokens) generated by the OIDC provider SHALL never be logged, echoed, or exposed in workflow logs or outputs.

> **Manual Setup Required — AWS OIDC Integration:**
> Before the CI workflows can authenticate to AWS, a repository administrator must complete the following one-time setup steps:
>
> 1. **Create the GitHub OIDC Identity Provider in AWS IAM:**
>    - In the AWS Console → IAM → Identity Providers, add a new OpenID Connect provider
>    - Provider URL: `https://token.actions.githubusercontent.com`
>    - Audience: `sts.amazonaws.com`
>
> 2. **Create a read-only IAM role for plan operations:**
>    - Create an IAM role with a trust policy allowing assumption by `token.actions.githubusercontent.com`
>    - Scope the trust policy condition to your repository: `"token.actions.githubusercontent.com:sub": "repo:<org>/<repo>:*"`
>    - Attach read-only policies sufficient for `terraform plan` (e.g. `ReadOnlyAccess` or a custom policy)
>    - Note the role ARN
>
> 3. **Store AWS identifiers as GitHub repository variables:**
>    - In the GitHub repository → Settings → Secrets and variables → Actions → Variables (not Secrets)
>    - Add a variable named `AWS_PLAN_ROLE_ARN` with the read-only role ARN (ARNs are identifiers, not credentials)
>    - Add a variable named `AWS_INIT_ROLE_ARN` with an identical or separate read-only role ARN for post-merge operations
>    - Add a variable named `AWS_REGION` with the target AWS region (e.g. `eu-west-1`)
>    - **Note:** Role ARNs are resource identifiers and are safe to store as variables. Temporary AWS credentials (STS access key, secret key, and session token) generated by OIDC are never stored and expire within minutes.

---

### Requirement 6: Unit Test Framework

**User Story:** As a developer, I want a unit test framework configured for Terraform modules, so that I can write and run automated tests that validate module logic without deploying real infrastructure.

#### Acceptance Criteria

1. THE Repository SHALL contain a `terraform/test/` directory housing all unit test files.
2. THE unit test framework SHALL use Terratest (Go-based) as the testing library.
3. THE Repository SHALL contain a `terraform/test/go.mod` and `terraform/test/go.sum` file with Terratest and all transitive dependencies pinned to exact versions (no range operators).
4. WHEN a developer runs `make test`, THE Unit_Test suite SHALL invoke `go test -v -timeout 30m ./...` (or equivalent) in the `terraform/test/` directory, stream output to stdout, and exit with a non-zero status code if any test fails.
5. THE Repository SHALL contain at least one example unit test file in `terraform/test/` with at least one exported test function named with a `Test` prefix that imports the Terratest library.
6. IF a unit test fails, THEN THE Unit_Test suite SHALL print the failing test name to stdout and, where the Terratest assertion library provides them, the expected and actual values, before exiting with a non-zero status code.

---

### Requirement 7: CloudFormation Backend Bootstrap

**User Story:** As a developer, I want a CloudFormation template that provisions the S3 bucket and DynamoDB table required by Terraform's remote state backend, so that I can create the backend infrastructure before running any Terraform commands.

#### Acceptance Criteria

1. THE Repository SHALL contain a `bootstrap/cloudformation.yml` file defining a CloudFormation template that provisions exactly two resources: an S3 bucket for Terraform state storage and a DynamoDB table for state locking.
2. THE S3 bucket resource in `bootstrap/cloudformation.yml` SHALL have versioning enabled, server-side encryption enabled (AES-256 or AWS KMS), and block all public access.
3. THE DynamoDB table resource in `bootstrap/cloudformation.yml` SHALL use `LockID` as the partition key with type `String` and SHALL use PAY_PER_REQUEST billing mode.
4. THE `bootstrap/cloudformation.yml` template SHALL accept the following CloudFormation parameters: `Environment` (e.g., `dev`, `prod`) and `ProjectName`, and SHALL use them to construct resource names (e.g., `<ProjectName>-<Environment>-terraform-state`).
5. THE `bootstrap/cloudformation.yml` template SHALL export the S3 bucket name and DynamoDB table name as CloudFormation stack outputs.
6. WHEN a developer runs `make bootstrap`, THE Makefile SHALL invoke `aws cloudformation deploy` with the stack name derived from the `ENV` variable, the template file at `bootstrap/cloudformation.yml`, and pass `Environment` and `ProjectName` as parameter overrides.
7. WHEN `make bootstrap` completes successfully, THE CloudFormation stack SHALL be in `CREATE_COMPLETE` or `UPDATE_COMPLETE` state and the S3 bucket and DynamoDB table SHALL exist in the target AWS account.

---

### Requirement 8: GitHub Repository Branch Protection

**User Story:** As a repository administrator, I want branch protection rules configured on the `main` branch, so that all changes are reviewed and pass CI checks before being merged.

#### Acceptance Criteria

1. THE `main` branch SHALL be configured with branch protection rules that require at least one pull request approval before merging.
2. THE `main` branch SHALL be configured to require the following status checks to pass before merging: `terraform fmt -check`, `tflint`, and `terraform plan`.
3. THE `main` branch SHALL be configured to dismiss stale pull request approvals when new commits are pushed.
4. THE `main` branch SHALL be configured to prevent direct pushes — all changes must be introduced via a pull request.
5. THE `main` branch SHALL be configured to require branches to be up to date with `main` before merging.

> **Note:** These branch protection rules are a manual GitHub repository configuration step and cannot be enforced via files committed to the repository. A repository administrator must configure these rules in the GitHub repository settings under Settings → Branches after the repository is created.
