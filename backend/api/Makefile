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

# Full AI Demo: Start Infra (LocalStack+AI), Setup Misconfigs, Run Test
e2e-ai-demo:
	@echo "=== 1. Starting Infrastructure (LocalStack, Postgres, Neo4j, AI Service) ==="
	@if [ -z "$(OPENAI_API_KEY)" ]; then echo "WARNING: OPENAI_API_KEY is not set. AI summaries might be mocked or fail."; fi
	cd ../../infra && docker compose up -d --build localstack postgres neo4j ai-service
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "=== 2. Setting up Misconfigurations ==="
	@./e2e/setup-misconfigs.sh
	@echo "=== 3. Running E2E Test with AI Summarization (Direct & GraphQL) ==="
	@echo "Look for 'Risk Score', 'Summary Text', and 'Commands' in the output below:"
	go test -v -timeout 10m -run "TestSecurityService_E2E|TestGraphQL_StartScan_E2E" ./e2e/...

# Stop the AI Demo Infrastructure
e2e-ai-down:
	cd ../../infra && docker compose down

# Check if LocalStack is running
e2e-status:
	@curl -s http://localhost:4566/_localstack/health | jq . || echo "LocalStack is not running"

# View LocalStack logs
e2e-logs:
	docker compose -f e2e/docker-compose.yml logs -f localstack

# Setup misconfigured AWS resources for testing (without starting LocalStack)
e2e-setup-only:
	@echo "Setting up misconfigured AWS resources in LocalStack..."
	@./e2e/setup-misconfigs.sh

# Full E2E workflow: start LocalStack, setup misconfigs, run tests
e2e-full: e2e-up e2e-setup-misconfigs
	@echo ""
	@echo "=== Running E2E Scanner Tests ==="
	go test -v -timeout 10m ./e2e/...

# Clean up and reset LocalStack
e2e-reset: e2e-down e2e-up
	@echo "LocalStack has been reset"
