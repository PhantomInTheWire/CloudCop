# CloudCop Onboarding

Deploy the IAM role for CloudCop security scans in your AWS account.

## Quick Setup

1. **Get the files**: Copy this directory to your local machine
2. **Update platform account ID** in `main.tf` (replace `123456789012`)
3. **Run Terraform**:
   ```bash
   terraform init
   terraform apply -var="external_id=YOUR_EXTERNAL_ID"
   ```
4. **Share the role ARN** from the output with CloudCop

## Variables

- `external_id`: Provided by CloudCop
- `platform_account_id`: CloudCop's AWS account ID

## What It Does

Creates an IAM role with read-only access that CloudCop can assume for security scanning. Uses external ID for security.