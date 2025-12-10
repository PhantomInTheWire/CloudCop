Infrastructure configurations and Docker files for the CloudCop monorepo.

## Docker Compose Environment

### Quickstart
```bash
cd infra
docker compose up
```

### Services

| Service        | Port | Purpose                     |
| -------------- | ---- | --------------------------- |
| postgres       | 5432 | Database                    |
| localstack     | 4566 | AWS emulator                |
| backend-api    | 8080 | Go GraphQL                  |
| backend-worker | —    | Background jobs             |
| ai-service     | 5001 | FastAPI + DSPy              |
| frontend       | 3000 | Next.js                     |

All services use environment variables from `.env` in repo root.

## 🧱 Infrastructure Base

This section contains templates for deploying CloudCop infrastructure in AWS.

### CloudCop Template

See [cloudcop-template/README.md](cloudcop-template/README.md) for Terraform modules that handle user onboarding and provide reusable infrastructure components.

### Terraform

See [terraform/README.md](terraform/README.md) for infrastructure as code using Terraform.