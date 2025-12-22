.PHONY: e2e-up e2e-down e2e-test e2e-test-s3 e2e-test-ec2 e2e-test-iam e2e-test-lambda e2e-test-dynamodb

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
