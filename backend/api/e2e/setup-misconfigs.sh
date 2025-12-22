#!/bin/bash
set -e

# CloudCop E2E Test Setup - Creates misconfigured AWS resources in LocalStack
# Usage: ./setup-misconfigs.sh [--endpoint URL]
#
# This script creates a variety of insecure AWS resource configurations
# that CloudCop scanners should detect. Use this to test the full scanning
# pipeline against realistic misconfigurations.

export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-east-1
ENDPOINT="${1:-http://localhost:4566}"

echo "=== CloudCop E2E Setup: Creating misconfigured AWS resources ==="
echo "Endpoint: $ENDPOINT"
echo ""

# ============================================================================
# S3 Misconfigurations
# ============================================================================
echo "[S3] Creating misconfigured buckets..."

# 1. Public bucket (CRITICAL - data exposure risk)
echo "  - Creating public-sensitive-data bucket (public-read ACL)..."
aws s3api create-bucket --bucket public-sensitive-data --endpoint-url $ENDPOINT 2>/dev/null || true
aws s3api put-bucket-acl --bucket public-sensitive-data --acl public-read --endpoint-url $ENDPOINT 2>/dev/null || echo "    (ACL may not be fully supported in LocalStack)"

# 2. Unencrypted bucket (HIGH - data at rest not protected)
echo "  - Creating unencrypted-bucket (no server-side encryption)..."
aws s3api create-bucket --bucket unencrypted-bucket --endpoint-url $ENDPOINT 2>/dev/null || true

# 3. No versioning (MEDIUM - no protection against accidental deletion)
echo "  - Creating no-versioning-bucket (versioning disabled)..."
aws s3api create-bucket --bucket no-versioning-bucket --endpoint-url $ENDPOINT 2>/dev/null || true

# 4. No logging (LOW - no audit trail)
echo "  - Creating no-logging-bucket (no access logging)..."
aws s3api create-bucket --bucket no-logging-bucket --endpoint-url $ENDPOINT 2>/dev/null || true

# 5. No lifecycle policy (LOW - no data retention management)
echo "  - Creating no-lifecycle-bucket (no lifecycle rules)..."
aws s3api create-bucket --bucket no-lifecycle-bucket --endpoint-url $ENDPOINT 2>/dev/null || true

# ============================================================================
# DynamoDB Misconfigurations
# ============================================================================
echo ""
echo "[DynamoDB] Creating misconfigured tables..."

# 1. Table without encryption (HIGH)
echo "  - Creating unencrypted-table (no encryption at rest)..."
aws dynamodb create-table \
    --table-name unencrypted-table \
    --attribute-definitions AttributeName=id,AttributeType=S \
    --key-schema AttributeName=id,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST \
    --endpoint-url $ENDPOINT 2>/dev/null || true

# 2. Table without PITR (MEDIUM - no point-in-time recovery)
echo "  - Creating no-pitr-table (point-in-time recovery disabled)..."
aws dynamodb create-table \
    --table-name no-pitr-table \
    --attribute-definitions AttributeName=id,AttributeType=S \
    --key-schema AttributeName=id,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST \
    --endpoint-url $ENDPOINT 2>/dev/null || true

# 3. Table without TTL (LOW - no automatic data expiration)
echo "  - Creating no-ttl-table (no TTL configured)..."
aws dynamodb create-table \
    --table-name no-ttl-table \
    --attribute-definitions AttributeName=id,AttributeType=S \
    --key-schema AttributeName=id,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST \
    --endpoint-url $ENDPOINT 2>/dev/null || true

# ============================================================================
# IAM Misconfigurations
# ============================================================================
echo ""
echo "[IAM] Creating overprivileged entities..."

# 1. User with full admin policy (CRITICAL)
echo "  - Creating admin-user with full admin permissions..."
aws iam create-user --user-name admin-user --endpoint-url $ENDPOINT 2>/dev/null || true

cat > /tmp/admin-policy.json << 'EOF'
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Action": "*",
        "Resource": "*"
    }]
}
EOF

aws iam create-policy \
    --policy-name AdminPolicy \
    --policy-document file:///tmp/admin-policy.json \
    --endpoint-url $ENDPOINT 2>/dev/null || true

aws iam attach-user-policy \
    --user-name admin-user \
    --policy-arn arn:aws:iam::000000000000:policy/AdminPolicy \
    --endpoint-url $ENDPOINT 2>/dev/null || true

# 2. Create access keys (for detecting old/unused keys)
echo "  - Creating access keys for admin-user..."
aws iam create-access-key --user-name admin-user --endpoint-url $ENDPOINT 2>/dev/null || true

# 3. User without MFA (HIGH - console access without MFA)
echo "  - Creating no-mfa-user with console access but no MFA..."
aws iam create-user --user-name no-mfa-user --endpoint-url $ENDPOINT 2>/dev/null || true
aws iam create-login-profile \
    --user-name no-mfa-user \
    --password "TempPassword123!" \
    --no-password-reset-required \
    --endpoint-url $ENDPOINT 2>/dev/null || true

# 4. User with inline policy (MEDIUM - policies should be managed)
echo "  - Creating inline-policy-user with inline policy..."
aws iam create-user --user-name inline-policy-user --endpoint-url $ENDPOINT 2>/dev/null || true

cat > /tmp/inline-policy.json << 'EOF'
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Action": ["s3:*", "ec2:*"],
        "Resource": "*"
    }]
}
EOF

aws iam put-user-policy \
    --user-name inline-policy-user \
    --policy-name InlineS3EC2Policy \
    --policy-document file:///tmp/inline-policy.json \
    --endpoint-url $ENDPOINT 2>/dev/null || true

# 5. Role with overly permissive trust policy (HIGH - cross-account risk)
echo "  - Creating overly-permissive-role with wide trust policy..."
cat > /tmp/permissive-trust.json << 'EOF'
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Principal": {"AWS": "*"},
        "Action": "sts:AssumeRole"
    }]
}
EOF

aws iam create-role \
    --role-name overly-permissive-role \
    --assume-role-policy-document file:///tmp/permissive-trust.json \
    --endpoint-url $ENDPOINT 2>/dev/null || true

# ============================================================================
# EC2 Security Group Misconfigurations
# ============================================================================
echo ""
echo "[EC2] Creating insecure security groups..."

# 1. Open SSH (CRITICAL - port 22 open to world)
echo "  - Creating open-ssh security group (port 22 to 0.0.0.0/0)..."
aws ec2 create-security-group \
    --group-name open-ssh \
    --description "Insecure: SSH open to 0.0.0.0/0" \
    --endpoint-url $ENDPOINT 2>/dev/null || true

SG_SSH=$(aws ec2 describe-security-groups \
    --group-names open-ssh \
    --query 'SecurityGroups[0].GroupId' \
    --output text \
    --endpoint-url $ENDPOINT 2>/dev/null)

if [ -n "$SG_SSH" ] && [ "$SG_SSH" != "None" ]; then
    aws ec2 authorize-security-group-ingress \
        --group-id $SG_SSH \
        --protocol tcp --port 22 --cidr 0.0.0.0/0 \
        --endpoint-url $ENDPOINT 2>/dev/null || true
fi

# 2. Open RDP (CRITICAL - port 3389 open to world)
echo "  - Creating open-rdp security group (port 3389 to 0.0.0.0/0)..."
aws ec2 create-security-group \
    --group-name open-rdp \
    --description "Insecure: RDP open to 0.0.0.0/0" \
    --endpoint-url $ENDPOINT 2>/dev/null || true

SG_RDP=$(aws ec2 describe-security-groups \
    --group-names open-rdp \
    --query 'SecurityGroups[0].GroupId' \
    --output text \
    --endpoint-url $ENDPOINT 2>/dev/null)

if [ -n "$SG_RDP" ] && [ "$SG_RDP" != "None" ]; then
    aws ec2 authorize-security-group-ingress \
        --group-id $SG_RDP \
        --protocol tcp --port 3389 --cidr 0.0.0.0/0 \
        --endpoint-url $ENDPOINT 2>/dev/null || true
fi

# 3. Open database ports (HIGH)
echo "  - Creating open-database security group (MySQL, PostgreSQL, MongoDB, Redis to 0.0.0.0/0)..."
aws ec2 create-security-group \
    --group-name open-database \
    --description "Insecure: Database ports open to 0.0.0.0/0" \
    --endpoint-url $ENDPOINT 2>/dev/null || true

SG_DB=$(aws ec2 describe-security-groups \
    --group-names open-database \
    --query 'SecurityGroups[0].GroupId' \
    --output text \
    --endpoint-url $ENDPOINT 2>/dev/null)

if [ -n "$SG_DB" ] && [ "$SG_DB" != "None" ]; then
    # MySQL (3306)
    aws ec2 authorize-security-group-ingress \
        --group-id $SG_DB \
        --protocol tcp --port 3306 --cidr 0.0.0.0/0 \
        --endpoint-url $ENDPOINT 2>/dev/null || true
    # PostgreSQL (5432)
    aws ec2 authorize-security-group-ingress \
        --group-id $SG_DB \
        --protocol tcp --port 5432 --cidr 0.0.0.0/0 \
        --endpoint-url $ENDPOINT 2>/dev/null || true
    # MongoDB (27017)
    aws ec2 authorize-security-group-ingress \
        --group-id $SG_DB \
        --protocol tcp --port 27017 --cidr 0.0.0.0/0 \
        --endpoint-url $ENDPOINT 2>/dev/null || true
    # Redis (6379)
    aws ec2 authorize-security-group-ingress \
        --group-id $SG_DB \
        --protocol tcp --port 6379 --cidr 0.0.0.0/0 \
        --endpoint-url $ENDPOINT 2>/dev/null || true
fi

# 4. All traffic open (CRITICAL)
echo "  - Creating all-traffic-open security group (all ports to 0.0.0.0/0)..."
aws ec2 create-security-group \
    --group-name all-traffic-open \
    --description "Insecure: All traffic open to 0.0.0.0/0" \
    --endpoint-url $ENDPOINT 2>/dev/null || true

SG_ALL=$(aws ec2 describe-security-groups \
    --group-names all-traffic-open \
    --query 'SecurityGroups[0].GroupId' \
    --output text \
    --endpoint-url $ENDPOINT 2>/dev/null)

if [ -n "$SG_ALL" ] && [ "$SG_ALL" != "None" ]; then
    aws ec2 authorize-security-group-ingress \
        --group-id $SG_ALL \
        --protocol -1 --cidr 0.0.0.0/0 \
        --endpoint-url $ENDPOINT 2>/dev/null || true
fi

# ============================================================================
# Lambda Misconfigurations
# ============================================================================
echo ""
echo "[Lambda] Creating insecure functions..."

# Create IAM role for Lambda execution
echo "  - Creating lambda-execution-role..."
cat > /tmp/lambda-trust-policy.json << 'EOF'
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Principal": {"Service": "lambda.amazonaws.com"},
        "Action": "sts:AssumeRole"
    }]
}
EOF

aws iam create-role \
    --role-name lambda-execution-role \
    --assume-role-policy-document file:///tmp/lambda-trust-policy.json \
    --endpoint-url $ENDPOINT 2>/dev/null || true

# Create a simple Lambda function code
cat > /tmp/insecure-lambda.py << 'EOF'
import os

def lambda_handler(event, context):
    # This function has secrets in environment variables - BAD PRACTICE!
    db_password = os.environ.get('DB_PASSWORD', '')
    api_key = os.environ.get('API_KEY', '')

    return {
        'statusCode': 200,
        'body': 'Hello from insecure Lambda!'
    }
EOF

cd /tmp && zip -q -o insecure-lambda.zip insecure-lambda.py

# 1. Lambda with secrets in environment variables (CRITICAL)
echo "  - Creating insecure-lambda (secrets in env vars, high timeout, no tracing)..."
aws lambda create-function \
    --function-name insecure-lambda \
    --runtime python3.11 \
    --role arn:aws:iam::000000000000:role/lambda-execution-role \
    --handler insecure-lambda.lambda_handler \
    --zip-file fileb:///tmp/insecure-lambda.zip \
    --environment "Variables={DB_PASSWORD=super_secret_password123,API_KEY=sk-1234567890abcdef,AWS_SECRET_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE}" \
    --timeout 300 \
    --endpoint-url $ENDPOINT 2>/dev/null || true

# 2. Lambda without DLQ (MEDIUM - failed invocations are lost)
echo "  - Creating no-dlq-lambda (no dead-letter queue)..."
aws lambda create-function \
    --function-name no-dlq-lambda \
    --runtime python3.11 \
    --role arn:aws:iam::000000000000:role/lambda-execution-role \
    --handler insecure-lambda.lambda_handler \
    --zip-file fileb:///tmp/insecure-lambda.zip \
    --timeout 60 \
    --endpoint-url $ENDPOINT 2>/dev/null || true

# 3. Lambda with excessive timeout (MEDIUM - potential cost/security issue)
echo "  - Creating high-timeout-lambda (15 minute timeout)..."
aws lambda create-function \
    --function-name high-timeout-lambda \
    --runtime python3.11 \
    --role arn:aws:iam::000000000000:role/lambda-execution-role \
    --handler insecure-lambda.lambda_handler \
    --zip-file fileb:///tmp/insecure-lambda.zip \
    --timeout 900 \
    --endpoint-url $ENDPOINT 2>/dev/null || true

# ============================================================================
# Cleanup temp files
# ============================================================================
rm -f /tmp/admin-policy.json /tmp/inline-policy.json /tmp/permissive-trust.json \
      /tmp/lambda-trust-policy.json /tmp/insecure-lambda.py /tmp/insecure-lambda.zip

# ============================================================================
# Summary
# ============================================================================
echo ""
echo "=== CloudCop E2E Setup Complete ==="
echo ""
echo "Created misconfigurations:"
echo ""
echo "  S3 Buckets (5):"
echo "    - public-sensitive-data    [CRITICAL] Public-read ACL"
echo "    - unencrypted-bucket       [HIGH]     No server-side encryption"
echo "    - no-versioning-bucket     [MEDIUM]   Versioning disabled"
echo "    - no-logging-bucket        [LOW]      No access logging"
echo "    - no-lifecycle-bucket      [LOW]      No lifecycle rules"
echo ""
echo "  DynamoDB Tables (3):"
echo "    - unencrypted-table        [HIGH]     No encryption at rest"
echo "    - no-pitr-table            [MEDIUM]   No point-in-time recovery"
echo "    - no-ttl-table             [LOW]      No TTL configured"
echo ""
echo "  IAM Users/Roles (5):"
echo "    - admin-user               [CRITICAL] Full admin policy + access keys"
echo "    - no-mfa-user              [HIGH]     Console access without MFA"
echo "    - inline-policy-user       [MEDIUM]   Uses inline policy instead of managed"
echo "    - overly-permissive-role   [HIGH]     Trust policy allows any AWS principal"
echo "    - lambda-execution-role    [OK]       Lambda execution role"
echo ""
echo "  EC2 Security Groups (4):"
echo "    - open-ssh                 [CRITICAL] Port 22 open to 0.0.0.0/0"
echo "    - open-rdp                 [CRITICAL] Port 3389 open to 0.0.0.0/0"
echo "    - open-database            [HIGH]     MySQL/PostgreSQL/MongoDB/Redis open"
echo "    - all-traffic-open         [CRITICAL] All ports open to 0.0.0.0/0"
echo ""
echo "  Lambda Functions (3):"
echo "    - insecure-lambda          [CRITICAL] Secrets in env vars, no tracing"
echo "    - no-dlq-lambda            [MEDIUM]   No dead-letter queue"
echo "    - high-timeout-lambda      [MEDIUM]   Excessive 15-minute timeout"
echo ""
echo "Run 'make e2e-test' to scan these resources with CloudCop scanners."
echo "Run 'make e2e-list-resources' to view all created resources."
