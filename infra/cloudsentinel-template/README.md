# CloudCop User Onboarding

This directory contains Terraform configuration to deploy the IAM role required for CloudCop to perform security scans in your AWS account.

## Prerequisites

- [Terraform](https://developer.hashicorp.com/terraform/downloads) installed
- AWS CLI configured with credentials to your AWS account
- Your CloudCop platform account ID and external ID

## Quick Start

1. **Clone or download these files** to a local directory:
   ```bash
   # If you have the CloudCop repository:
   cp -r infra/user-onboarding ~/cloudcop-onboarding
   cd ~/cloudcop-onboarding
   ```

2. **Update the platform account ID** in `main.tf`:
   - Open `main.tf` and replace `123456789012` on line 18 with your actual CloudCop platform account ID

3. **Initialize Terraform**:
   ```bash
   terraform init
   ```

4. **Deploy the infrastructure**:
   ```bash
   terraform apply -var="external_id=YOUR_EXTERNAL_ID_HERE"
   ```
   - Replace `YOUR_EXTERNAL_ID_HERE` with the external ID provided by CloudCop
   - Review the plan and type `yes` when prompted

5. **Provide the role ARN** to CloudCop:
   - After deployment, copy the `role_arn` value from the output
   - Share this ARN with your CloudCop representative

## Configuration

### Required Variables

- `external_id`: A unique identifier provided by CloudCop for security
- `platform_account_id`: The AWS account ID of the CloudCop platform (default: 123456789012)

### Optional: Using a terraform.tfvars file

Create a `terraform.tfvars` file to store your variables:

```hcl
external_id = "your-unique-external-id-here"
platform_account_id = "your-cloudcop-account-id"
```

Then run:
```bash
terraform apply
```

## What Gets Created

This Terraform configuration creates:

1. **IAM Role**: `GuardSecurityScanRole`
   - Allows the CloudCop platform to assume this role
   - Requires external ID validation for security
   - Has read-only access to your AWS resources

2. **Policy Attachment**: ReadOnlyAccess
   - Grants read permissions to scan your AWS resources
   - Cannot modify or delete any resources

## Security Considerations

- The role uses an external ID for additional security
- Only has read-only permissions (ReadOnlyAccess policy)
- Can only be assumed by the specified CloudCop platform account
- All actions are logged in AWS CloudTrail

## Cleanup

To remove the IAM role when no longer needed:

```bash
terraform destroy
```

## Support

If you encounter any issues during deployment:

1. Verify your AWS credentials are properly configured
2. Ensure you have the necessary IAM permissions to create roles
3. Contact your CloudCop representative for assistance

## Files Description

- `main.tf`: Main Terraform configuration with IAM role definition
- `variables.tf`: Input variables (external_id, platform_account_id)
- `outputs.tf`: Output values (role_arn, role_name)
- `README.md`: This documentation file