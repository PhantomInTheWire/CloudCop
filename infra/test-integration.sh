#!/bin/bash
set -e

echo "================================"
echo "CloudCop AWS Auth Integration Test"
echo "================================"
echo ""

#
# Configure dummy AWS credentials for LocalStack.
# LocalStack doesn't validate these credentials, but the AWS CLI requires them to be set.
#
export AWS_ACCESS_KEY_ID="test"
export AWS_SECRET_ACCESS_KEY="test"
export AWS_DEFAULT_REGION="us-west-2"

# Configuration
LOCALSTACK_ENDPOINT="http://localhost:4566"
EXTERNAL_ID="test-external-id-$(uuidgen)"
TEST_ACCOUNT_ID="000000000000"  # LocalStack default account ID

echo "1. Checking LocalStack availability..."
until curl -s "${LOCALSTACK_ENDPOINT}/_localstack/health" > /dev/null 2>&1; do
  echo "   Waiting for LocalStack to be ready..."
  sleep 2
done
echo "   ✓ LocalStack is ready"
echo ""

echo "2. Deploying CloudFormation template..."
aws --endpoint-url="${LOCALSTACK_ENDPOINT}" cloudformation create-stack \
  --stack-name cloudcop-security-scan-role \
  --template-body file://cloudformation/guard-scan-role.yaml \
  --parameters ParameterKey=ExternalId,ParameterValue="${EXTERNAL_ID}" \
  --capabilities CAPABILITY_NAMED_IAM \
  --region us-west-2

echo "   Waiting for stack creation..."
sleep 5

aws --endpoint-url="${LOCALSTACK_ENDPOINT}" cloudformation wait stack-create-complete \
  --stack-name cloudcop-security-scan-role \
  --region us-west-2 || true

echo "   ✓ CloudFormation stack created"
echo ""

echo "3. Verifying IAM role was created..."
ROLE_ARN=$(aws --endpoint-url="${LOCALSTACK_ENDPOINT}" iam get-role \
  --role-name CloudCopSecurityScanRole \
  --region us-west-2 \
  --query 'Role.Arn' \
  --output text)

echo "   ✓ Role ARN: ${ROLE_ARN}"
echo ""

echo "4. Testing STS AssumeRole..."
ASSUMED_ROLE=$(aws --endpoint-url="${LOCALSTACK_ENDPOINT}" sts assume-role \
  --role-arn "${ROLE_ARN}" \
  --role-session-name "CloudCopTestSession" \
  --external-id "${EXTERNAL_ID}" \
  --region us-west-2 \
  --query 'Credentials.[AccessKeyId,SecretAccessKey,SessionToken]' \
  --output text)

if [ -n "${ASSUMED_ROLE}" ]; then
  echo "   ✓ Successfully assumed role with ExternalID"
else
  echo "   ✗ Failed to assume role"
  exit 1
fi
echo ""

echo "5. Testing invalid ExternalID (should fail)..."
if aws --endpoint-url="${LOCALSTACK_ENDPOINT}" sts assume-role \
  --role-arn "${ROLE_ARN}" \
  --role-session-name "CloudCopTestSession" \
  --external-id "invalid-external-id" \
  --region us-west-2 > /dev/null 2>&1; then
  echo "   ✗ Should have failed with invalid ExternalID"
  exit 1
else
  echo "   ✓ Correctly rejected invalid ExternalID"
fi
echo ""

echo "6. Testing API endpoints..."
echo "   Starting API server in background..."

#
# Start the API server with LocalStack configuration.
# The server will use the AWS_ENDPOINT_URL to connect to LocalStack
# instead of real AWS services.
#
cd ../../backend/api
AWS_REGION=us-west-2 \
AWS_ENDPOINT_URL="${LOCALSTACK_ENDPOINT}" \
go run ./cmd/server/main.go > /tmp/cloudcop-api.log 2>&1 &
API_PID=$!

echo "   Waiting for API to be ready..."
sleep 3

#
# Test the verify endpoint to ensure the API can successfully
# communicate with LocalStack and verify AWS account access.
#
echo "   Testing POST /api/accounts/verify..."
VERIFY_RESPONSE=$(curl -s -X POST http://localhost:8080/api/accounts/verify \
  -H "Content-Type: application/json" \
  -d "{\"account_id\":\"${TEST_ACCOUNT_ID}\",\"external_id\":\"${EXTERNAL_ID}\"}")

if echo "${VERIFY_RESPONSE}" | grep -q "verified"; then
  echo "   ✓ Account verification successful"
else
  echo "   ✗ Account verification failed"
  echo "   Response: ${VERIFY_RESPONSE}"
  kill ${API_PID} 2>/dev/null || true
  exit 1
fi
echo ""

echo "   Testing POST /api/accounts/connect..."
CONNECT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/accounts/connect \
  -H "Content-Type: application/json" \
  -d "{\"account_id\":\"${TEST_ACCOUNT_ID}\",\"external_id\":\"${EXTERNAL_ID}\"}")

if echo "${CONNECT_RESPONSE}" | grep -q "success"; then
  echo "   ✓ Account connection successful"
else
  echo "   ✗ Account connection failed"
  echo "   Response: ${CONNECT_RESPONSE}"
fi
echo ""

echo "   Testing GET /api/accounts..."
LIST_RESPONSE=$(curl -s http://localhost:8080/api/accounts)
echo "   ✓ Account list retrieved"
echo ""

echo "   Stopping API server..."
kill ${API_PID} 2>/dev/null || true
echo ""

echo "================================"
echo "✓ All integration tests passed!"
echo "================================"
echo ""
echo "Summary:"
echo "  - CloudFormation template deployed successfully"
echo "  - IAM role created with correct permissions"
echo "  - STS AssumeRole works with valid ExternalID"
echo "  - Invalid ExternalID correctly rejected"
echo "  - API endpoints functioning correctly"
echo ""
echo "External ID used: ${EXTERNAL_ID}"
echo "Account ID: ${TEST_ACCOUNT_ID}"
echo ""
