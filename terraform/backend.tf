# Remote state backend configuration for S3 + DynamoDB
#
# Backend values are NOT hardcoded in this file. Instead, they are injected
# via `-backend-config` flags at init time, using the per-environment
# `backend.tfvars` file (e.g., environments/dev/backend.tfvars).
#
# This pattern allows a single backend.tf definition to support multiple
# environments without duplicating or committing sensitive AWS account
# identifiers to the repository.
#
# Usage:
#   terraform init -backend-config=environments/dev/backend.tfvars
#   terraform init -backend-config=environments/prod/backend.tfvars

terraform {
  backend "s3" {
    # All values supplied via -backend-config flag
  }
}
