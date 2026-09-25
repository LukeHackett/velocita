# Velocita

Velocita is a serverless DORA-metrics pipeline for GitHub Actions. 

Velocita captures CI/CD deployment events and computes Deployment Frequency, Lead Time for Changes, and Change Failure Rate using an API Gateway → Lambda → Firehose → S3 → Athena data path.

## Directory Structure

```
velocita/
├── bootstrap/                    # CloudFormation templates for one-time setup
│   └── cloudformation.yml        # S3 state bucket + DynamoDB lock table
├── terraform/                    # Terraform configuration root
│   ├── backend.tf                # S3 backend declaration
│   ├── versions.tf               # Provider version constraints
│   ├── modules/                  # Reusable Terraform modules
│   ├── environments/             # Environment-specific configurations
│   │   ├── dev/                  # Development environment
│   │   │   ├── backend.tfvars    # Backend config for dev
│   │   │   └── terraform.tfvars  # Variable overrides for dev
│   │   └── prod/                 # Production environment
│   │       ├── backend.tfvars    # Backend config for prod
│   │       └── terraform.tfvars  # Variable overrides for prod
│   └── test/                     # Terratest unit tests
├── .github/                      # GitHub configuration
│   ├── actions/                  # Reusable composite actions
│   ├── workflows/                # CI/CD workflows
│   └── dependabot.yml            # Automated dependency updates
├── Makefile                      # Developer workflow commands
└── .tflint.hcl                   # TFLint configuration
```

### Key Directories

- **`terraform/modules/`**: Contains reusable Terraform modules. Each module is a self-contained unit of infrastructure code that can be shared across environments. Modules are populated by feature specs as the project evolves.

- **`terraform/environments/`**: Contains environment-specific configurations. Each environment (`dev`, `prod`) has its own directory with `backend.tfvars` for remote state configuration and `terraform.tfvars` for variable overrides. This separation ensures environment isolation and prevents cross-environment state collisions.

## Prerequisites

Install the following tools before contributing:

| Tool | Purpose | Installation |
|------|---------|--------------|
| [Terraform](https://developer.hashicorp.com/terraform/downloads) | Infrastructure as Code | `brew install terraform` |
| [AWS CLI](https://aws.amazon.com/cli/) | AWS operations | `brew install awscli` |
| [TFLint](https://github.com/terraform-linters/tflint) | Terraform linter | `brew install tflint` |
| [Trivy](https://github.com/aquasecurity/trivy) | Security scanner | `brew install trivy` |
| [Go](https://golang.org/) | Terratest framework | `brew install go` |
| [jq](https://stedolan.github.io/jq/) | JSON processing | `brew install jq` |

## Contributor Onboarding

### 1. Verify Dependencies

Run the dependency check to ensure all required tools are installed:

```bash
make check-deps
```

This command verifies that `terraform`, `aws`, `tflint`, `trivy`, `go`, and `jq` are present on PATH. The pinned Terraform version is tracked in `.tool-versions` for use with version managers like `asdf`.

### 2. Bootstrap Backend Infrastructure (One-Time Setup)

Before running any Terraform commands, bootstrap the remote state backend for your target environment:

```bash
make bootstrap ENV=dev
```

This deploys a CloudFormation stack that creates:
- S3 bucket for Terraform state storage
- DynamoDB table for state locking

### 3. Initialize Terraform

Initialize Terraform with the environment-specific backend configuration:

```bash
make init ENV=dev
```

This command:
- Initializes the Terraform working directory
- Configures the S3 backend using `terraform/environments/dev/backend.tfvars`
- Downloads provider plugins
- Generates the `.terraform.lock.hcl` file

## Development Workflow

### Running Terraform Commands

All Terraform operations use the `ENV` variable to select the target environment:

```bash
# Plan changes for dev environment
make plan ENV=dev

# Apply the generated plan
make apply ENV=dev

# Format all Terraform files
make fmt

# Run linter
make lint

# Scan for security vulnerabilities
make trivy
```

### Running Tests

The test suite uses [Terratest](https://github.com/gruntwork-io/terratest) for infrastructure validation:

```bash
make test
```

This runs all tests in the `terraform/test/` directory with a 30-minute timeout.

### Cleaning Up

To remove all auto-generated files (`.terraform/` directories, plan files, lock files, and Go build artifacts) and reset the project to a clean state:

```bash
make clean
```

Re-run `make init ENV=<env>` afterward to reinitialize.

### Available Make Targets

Run `make help` to see all available commands:

```bash
make help
```

## CI/CD Workflows

### Pre-Merge Workflow

Triggered on pull requests targeting `main`. Runs in parallel:
- **Lint job**: `terraform fmt` + `tflint`
- **Security job**: `trivy config` scan
- **Plan job**: `terraform plan` with artifact storage

Results are aggregated into a single PR comment.

### Post-Merge Workflow

Triggered on push to `main`. Runs:
- AWS OIDC authentication
- `terraform init`
- Terratest suite

**Note**: `terraform apply` is never run automatically. All applies are manual via `make apply ENV=<environment>`.

## AWS OIDC Authentication

CI workflows authenticate to AWS using GitHub OIDC (OpenID Connect). This eliminates the need for long-lived AWS credentials stored as GitHub secrets.

### Required Setup

Before CI workflows can run, a repository administrator must:

1. Create a GitHub OIDC Identity Provider in AWS IAM
2. Create IAM roles with appropriate permissions for plan and init operations
3. Store role ARNs as GitHub repository variables (not secrets):
   - `AWS_PLAN_ROLE_ARN` - Read-only role for pre-merge plan operations
   - `AWS_INIT_ROLE_ARN` - Read-only role for post-merge init operations
   - `AWS_REGION` - Target AWS region (e.g., `eu-west-1`)

See the [design document](.kiro/specs/terraform-project-bootstrap/design.md) for detailed OIDC setup instructions.

## License

MIT License - see [LICENSE](LICENSE) for details.
