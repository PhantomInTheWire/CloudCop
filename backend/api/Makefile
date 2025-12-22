.PHONY: e2e-up e2e-down e2e-test e2e-test-s3 e2e-test-ec2 e2e-test-iam e2e-test-lambda e2e-test-dynamodb e2e-setup-misconfigs e2e-list-resources e2e-full

# Start LocalStack for E2E testing
e2e-up:
	docker compose -f e2e/docker-compose.yml up -d --wait
	@echo "LocalStack is ready at http://localhost:4566"

# Stop LocalStack
e2e-down:
	docker compose -f e2e/docker-compose.yml down -v

# Run all E2E tests
e2e-test: e2e-up
	go test -v -timeout 10m ./e2e/... || ($(MAKE) e2e-down && exit 1)

# Run S3 E2E tests
e2e-test-s3: e2e-up
	go test -v -timeout 5m -run TestS3 ./e2e/...

# Run EC2 E2E tests
e2e-test-ec2: e2e-up
	go test -v -timeout 5m -run TestEC2 ./e2e/...

# Run IAM E2E tests
e2e-test-iam: e2e-up
	go test -v -timeout 5m -run TestIAM ./e2e/...

# Run Lambda E2E tests
e2e-test-lambda: e2e-up
	go test -v -timeout 5m -run TestLambda ./e2e/...

# Run DynamoDB E2E tests
e2e-test-dynamodb: e2e-up
	go test -v -timeout 5m -run TestDynamoDB ./e2e/...

# Run E2E tests without managing LocalStack (assume it's already running)
e2e-test-only:
	go test -v -timeout 10m ./e2e/...

# Check if LocalStack is running
e2e-status:
	@curl -s http://localhost:4566/_localstack/health | jq . || echo "LocalStack is not running"

# View LocalStack logs
e2e-logs:
	docker compose -f e2e/docker-compose.yml logs -f localstack

# Setup misconfigured AWS resources for testing
e2e-setup-misconfigs: e2e-up
	@echo "Setting up misconfigured AWS resources in LocalStack..."
	@./e2e/setup-misconfigs.sh

# List all resources created in LocalStack
e2e-list-resources:
	@echo ""
	@echo "=== S3 Buckets ==="
	@AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_DEFAULT_REGION=us-east-1 aws --endpoint-url http://localhost:4566 s3 ls 2>/dev/null || echo "  (none or LocalStack not running)"
	@echo ""
	@echo "=== DynamoDB Tables ==="
	@AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_DEFAULT_REGION=us-east-1 aws --endpoint-url http://localhost:4566 dynamodb list-tables --output table 2>/dev/null || echo "  (none or LocalStack not running)"
	@echo ""
	@echo "=== IAM Users ==="
	@AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_DEFAULT_REGION=us-east-1 aws --endpoint-url http://localhost:4566 iam list-users --query 'Users[*].UserName' --output table 2>/dev/null || echo "  (none or LocalStack not running)"
	@echo ""
	@echo "=== IAM Roles ==="
	@AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_DEFAULT_REGION=us-east-1 aws --endpoint-url http://localhost:4566 iam list-roles --query 'Roles[*].RoleName' --output table 2>/dev/null || echo "  (none or LocalStack not running)"
	@echo ""
	@echo "=== EC2 Security Groups ==="
	@AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_DEFAULT_REGION=us-east-1 aws --endpoint-url http://localhost:4566 ec2 describe-security-groups --query 'SecurityGroups[*].{Name:GroupName,ID:GroupId}' --output table 2>/dev/null || echo "  (none or LocalStack not running)"
	@echo ""
	@echo "=== Lambda Functions ==="
	@AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_DEFAULT_REGION=us-east-1 aws --endpoint-url http://localhost:4566 lambda list-functions --query 'Functions[*].FunctionName' --output table 2>/dev/null || echo "  (none or LocalStack not running)"

# Full E2E workflow: start LocalStack, setup misconfigs, run tests
e2e-full: e2e-up e2e-setup-misconfigs
	@echo ""
	@echo "=== Running E2E Scanner Tests ==="
	go test -v -timeout 10m ./e2e/...

# Clean up and reset LocalStack
e2e-reset: e2e-down e2e-up
	@echo "LocalStack has been reset"
