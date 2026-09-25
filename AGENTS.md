# AGENTS.md - Project Standards & Guidelines

## Project Overview

**Velocita** is a serverless DORA-metrics pipeline for GitHub Actions. It captures CI/CD deployment events and computes Deployment Frequency, Lead Time for Changes, and Change Failure Rate using an API Gateway → Lambda → Firehose → S3 → Athena data path.

Key characteristics:
- Fully serverless (no infrastructure maintenance)
- Pay-per-use cost model (~$0.50/month for standard workloads)
- GitHub OIDC authentication (zero long-lived credentials)
- Event-driven architecture with on-demand analytics
- Modular, spec-driven development approach

## Spec-Driven Development

**This project uses Kiro specs as the single source of truth for all requirements and design decisions.**

All changes—whether infrastructure, code, configuration, or documentation—must flow through the Kiro spec system:

1. **Requirements Phase** (`.kiro/specs/*/requirements.md`)
   - Define *what* the feature must do
   - Document acceptance criteria
   - Specify constraints and assumptions

2. **Design Phase** (`.kiro/specs/*/design.md`)
   - Define *how* the feature is implemented
   - Document architecture and component interactions
   - Explain design decisions and trade-offs

3. **Implementation Phase** (`.kiro/specs/*/tasks.md`)
   - Break design into discrete, testable tasks
   - Execute tasks incrementally
   - Verify each task against spec criteria

**No changes should be made to the codebase outside of a Kiro spec task.** This ensures:
- All work is traceable to documented requirements
- Design decisions are preserved for future reference
- Code changes are validated against acceptance criteria
- Implementation stays consistent with architecture

## Terraform Standards

### Structure
- **Single `terraform/` root** with shared `versions.tf`
- **Per-environment configs** in `terraform/environments/{env}/` (currently `dev`, `prod`)
- **Modules** in `terraform/modules/` for reusable infrastructure units
- **Backend** via S3 + DynamoDB (bootstrapped by CloudFormation, not Terraform)

### Version Pinning
- **Provider versions**: Use pessimistic constraints in `versions.tf` (e.g., `~> 5.0` for AWS provider)

### Provider Configuration
```hcl
# terraform/versions.tf - Example
terraform {
  required_version = "~> 1.9"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}
```

## Code Quality & Tooling

### Linting
- **tflint configuration** (`.tflint.hcl`) enables `terraform` plugin (recommended preset) and `aws` plugin
- **AWS-specific rules** catch deprecated resources, invalid patterns, compliance issues
- Run `make lint` locally; non-compliance blocks PRs (CI)

### Security Scanning
- **trivy config** scans Terraform files for HIGH/CRITICAL security vulnerabilities
- Run `make trivy` locally; security issues block PRs (CI)
- Exceptions stored in `.trivyignore.yaml` at repo root

### Testing
- **Framework**: Terratest (Go-based property testing)
- **Location**: `terraform/test/` directory
- **Scope**: Validates configuration validity, consistency, and format—does NOT deploy real AWS resources
- **Execution**: `make test` (30-minute timeout)

### Code Format
- **Indentation**: 2 spaces (HCL, YAML, JSON)
- **Line endings**: LF (enforced by `.gitattributes`)
- **EditorConfig**: `.editorconfig` provides IDE defaults

## CI/CD Patterns

### Pre-Merge Workflow (Pull Requests)
Triggered on `pull_request` events targeting `main`:
- **Parallel execution** of three validation jobs:
  - `lint` job: `terraform fmt -check` + `tflint`
  - `security` job: `trivy config` with HIGH/CRITICAL severity filter
  - `plan` job: `terraform plan` with artifact storage (1-day retention)
- **Sequential aggregation** via `comment` job:
  - Depends on all three validation jobs
  - Posts single PR comment with aggregated results
  - Fails workflow if any upstream job failed (prevents merge)

**Note**: Pre-merge workflow never runs `terraform apply`.

### Post-Merge Workflow (Push to Main)
Triggered on `push` to `main`:
- **Sequential verification** (halts on first failure):
  - AWS OIDC authentication (read-only role)
  - `terraform init` with backend configuration
  - Terratest suite execution (`go test -v -timeout 30m ./...`)

**Note**: Post-merge workflow also never runs `terraform apply`.

### Authentication
- **GitHub OIDC**: No long-lived AWS credentials stored in GitHub
- **Role assumption**: Temporary STS credentials valid for workflow duration
- **Session tracking**: `role-session-name` includes `github.run_id` for CloudTrail auditability
- **Repository variables**: Role ARNs and region stored as GitHub repository variables (non-secrets)

### Apply Strategy
- **Manual only**: All `terraform apply` operations are manual via `make apply ENV=<env>`
- **Plan artifact reuse**: Pre-merge plan is stored as artifact; can be applied without re-planning
- **Plan file guard**: `make apply` verifies plan file exists before executing

## Git & Commit Standards

### Branching
- Work in feature branches off `main`
- Branch naming: descriptive, lowercase with hyphens (e.g., `feature/api-gateway-setup`)
- All changes via pull request to `main`

### Commits
- Descriptive commit messages (imperative mood: "Add X" not "Added X")
- Reference Kiro spec task IDs when applicable
- Example: `Add S3 lifecycle rules (Task 1.2)`

### Pull Requests
- All CI checks must pass before merge
- Require at least one approval
- Branch protection: require branches to be up-to-date with `main`
- Stale approvals dismissed when new commits pushed

## Developer Workflow

### First-Time Setup
```bash
# 1. Verify dependencies
make check-deps

# 2. Bootstrap backend infrastructure (one-time per environment)
make bootstrap ENV=dev

# 3. Initialize Terraform
make init ENV=dev
```

### Daily Development
```bash
# Make changes to Terraform code in terraform/

# Check formatting and linting
make fmt
make lint

# Check security
make trivy

# Plan changes
make plan ENV=dev

# Review plan output, then apply
make apply ENV=dev

# Run test suite
make test

# Clean up when done (removes build artifacts)
make clean
```

### Makefile Targets
| Target | Purpose | Requires ENV? |
|--------|---------|---------------|
| `make check-deps` | Verify tools installed | No |
| `make bootstrap ENV=<env>` | Deploy CloudFormation backend stack | Yes |
| `make init ENV=<env>` | Initialize Terraform with backend config | Yes |
| `make plan ENV=<env>` | Run terraform plan and store output | Yes |
| `make apply ENV=<env>` | Apply stored plan file | Yes |
| `make fmt` | Format all Terraform files | No |
| `make lint` | Run tflint on all Terraform files | No |
| `make trivy` | Scan for security vulnerabilities | No |
| `make test` | Run Terratest suite | No |
| `make clean` | Remove all auto-generated files | No |
| `make help` | Display all targets with descriptions | No |

## Naming Conventions

### AWS Resources
Pattern: `{ProjectName}-{Environment}-{ResourceType}`

Examples:
- S3 bucket: `velocita-dev-terraform-state`
- DynamoDB table: `velocita-dev-terraform-locks`
- CloudFormation stack: `velocita-dev-bootstrap`
- Terraform state key: `dev/terraform.tfstate`

### Terraform Modules
- Lowercase with hyphens: `api-gateway-integration`, `lambda-authorizer`, etc.
- Descriptive names reflecting responsibility
- Located in `terraform/modules/`

### Terraform Variables & Outputs
- Snake_case: `aws_region`, `environment_name`, `resource_count`
- Descriptive: avoid abbreviations
- Include type hints in variable descriptions

## File Structure Conventions

```
velocita/
├── terraform/                          # Terraform root
│   ├── versions.tf                     # Provider version constraints
│   ├── backend.tf                      # S3 backend declaration
│   ├── modules/                        # Reusable modules
│   ├── environments/
│   │   ├── dev/
│   │   │   ├── backend.tfvars          # Backend config (non-secret)
│   │   │   └── terraform.tfvars        # Variable overrides
│   │   └── prod/
│   │       ├── backend.tfvars
│   │       └── terraform.tfvars
│   └── test/                           # Terratest suite
│       ├── go.mod
│       ├── go.sum
│       └── *_test.go
├── bootstrap/
│   └── cloudformation.yml              # One-time backend setup
├── .github/
│   ├── actions/                        # Reusable composite actions
│   │   ├── aws-oidc-auth/
│   │   └── setup-terraform/
│   ├── workflows/                      # CI/CD workflows
│   │   ├── pre-merge.yml
│   │   └── post-merge.yml
│   └── dependabot.yml                  # Automated dependency updates
├── .tflint.hcl                         # TFLint configuration
├── .trivyignore.yaml                   # Trivy ignore rules
├── Makefile                            # Developer commands
├── README.md                           # User-facing documentation
└── AGENTS.md                           # This file
```

## Testing Strategy

### What We Test
- **Configuration validity**: `terraform validate` succeeds
- **Backend completeness**: All required keys present, no credentials hardcoded
- **Lock file presence**: `.terraform.lock.hcl` exists and contains expected providers

### What We Don't Test
- Real AWS resource provisioning (integration tests deferred to later specs)
- State modifications or data transformations
- Multi-environment behavior (each environment tested in isolation)

### Running Tests
```bash
make test
```

Output streams to stdout; test suite exits non-zero on failure.

## Troubleshooting

### Terraform Lock File Stuck
```bash
terraform force-unlock <lock-id>
```
Use when a workflow terminates abnormally with a held DynamoDB lock.

### Plan File Not Found
Ensure you've run `make plan ENV=dev` before `make apply ENV=dev`:
```bash
make plan ENV=dev      # Creates tfplan-dev
make apply ENV=dev     # Uses tfplan-dev
```

### Dependency Check Failing
Verify all required binaries are installed:
```bash
make check-deps
```
Install missing tools as indicated. The pinned Terraform version is tracked in `.tool-versions`.

## Key Decisions & Rationale

### Why CloudFormation for Bootstrap?
Terraform cannot manage its own remote state backend with Terraform itself—you need a bootstrapping mechanism. CloudFormation is AWS-native, idempotent, and avoids a circular dependency.

### Why GitHub OIDC?
Eliminates long-lived AWS credential storage in GitHub secrets. Temporary STS credentials are scoped to the workflow duration and CloudTrail-auditabile via session name.

### Why Property-Based Tests?
Catches invariant violations (e.g., missing required keys) early and efficiently. Faster feedback than integration tests, and no need for AWS account access during PR validation.

### Why Per-Environment Configs in Code?
Non-secret values (S3 bucket names, regions, state file keys) are explicitly defined and reviewable. This prevents environment misconfigurations and documents the intended deployment topology.

### Why Manual Apply?
Automatic applies on merge introduce risk: a single bad PR could cascade through main branch into production. Manual `make apply` adds a human checkpoint—the developer reviews the plan one more time before committing changes to live infrastructure.

### Why GitHub Actions CI Only?
Centralizing all code quality checks (lint, security, formatting) into GitHub Actions CI/CD eliminates the overhead of maintaining dual local pre-commit hooks and CI checks. Developers run `make fmt`, `make lint`, and `make trivy` locally before pushing; CI enforces the same checks before merge. This reduces operational complexity and ensures consistent tooling versions across environments.

## Contributing

All contributions must follow this flow:

1. **Check the Kiro spec** in `.kiro/specs/` for requirements and design
2. **Create a branch** off `main`
3. **Make changes aligned with spec tasks**
4. **Run local validation**: `make fmt`, `make lint`, `make trivy`, `make test`
5. **Commit with descriptive messages**
6. **Open a PR** targeting `main`
7. **Wait for CI checks** to pass and approvals
8. **Merge** when ready

If requirements or design need adjustment, **open an issue or discussion proposing spec updates**—do not modify code before specs are updated.

## Resources

- **Terraform Docs**: https://www.terraform.io/docs
- **AWS Provider Docs**: https://registry.terraform.io/providers/hashicorp/aws/latest/docs
- **Terratest Guide**: https://github.com/gruntwork-io/terratest
- **TFLint Rules**: https://www.tflint.io/docs/rules/
- **GitHub Actions**: https://docs.github.com/en/actions
