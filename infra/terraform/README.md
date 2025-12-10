# Terraform Configuration

This directory contains Terraform configuration for deploying CloudCop infrastructure.

## Local Testing with LocalStack

To test locally with LocalStack:

1. Install LocalStack and start it:
   ```bash
   pip install localstack
   localstack start
   ```

2. Set environment variables:
   ```bash
   export AWS_ACCESS_KEY_ID=test
   export AWS_SECRET_ACCESS_KEY=test
   export AWS_DEFAULT_REGION=us-east-1
   ```

3. Initialize and apply:
   ```bash
   terraform init
   terraform plan -var="external_id=test-external-id"
   terraform apply -var="external_id=test-external-id"
   ```

## Production Deployment

For production deployment to AWS, ensure your AWS credentials are configured and update the provider account ID in `main.tf`.