# CloudCop

AI-powered cloud security tool that scans AWS environments for common misconfigurations and vulnerabilities. It provides actionable insights and even command-level fixes using LLMs to enhance cloud security with minimal setup.

## Features

- **AI-Powered Security Platform**: Built with Go (Gin) and Python (FastAPI) backends, and React.js/Next.js frontend
- **AWS Service Scanners**: Comprehensive security analysis for IAM, EC2, S3, Lambda, DynamoDB, and ECS
- **LLM-Based Remediation**: Automated command-line fixes for misconfigurations using DSPy for robust structures on unreliable LLM outputs
- **GraphQL API**: Real-time security scanning and vulnerability management with PostgreSQL database
- **Docker Deployment**: Multi-environment support (development/production)
- **Extensible Architecture**: Supports multiple cloud providers with role-based access control
- **SaaS Functionality**: Authentication middleware and Stripe payment webhook integration

## Quick Start

```bash
cd infra
docker compose up
```

Services will be available at:
- Frontend: http://localhost:3000
- API: http://localhost:8080
- AI: http://localhost:5001
- Worker: Logs in Docker Compose output

## Developer Setup

### Prerequisites

- Go 1.23+
- Node.js 20+
- Python 3.11+
- Docker
- pre-commit framework

### Setting Up Pre-commit Hooks

This project uses pre-commit hooks to catch issues before commits are made. The hooks mirror all CI checks for faster feedback.

**Automated Setup:**
```bash
./scripts/setup-dev.sh
```

**Manual Setup:**
```bash
# Install pre-commit
pip install pre-commit

# Install required tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Install Python tools
cd backend/ai
uv venv
uv pip install ruff mypy bandit
cd ../..

# Install commitlint
npm install -g @commitlint/cli @commitlint/config-conventional

# Install hooks
pre-commit install
pre-commit install --hook-type commit-msg
```

### Commit Message Format

This project enforces [Conventional Commits](https://www.conventionalcommits.org/) format:

```
type(scope): description

Examples:
  feat(backend): add new security scanner
  fix(frontend): resolve dashboard display issue
  chore(ci): update pre-commit hooks
  docs(readme): improve setup instructions
```

**Valid types:** `feat`, `fix`, `chore`, `docs`, `style`, `refactor`, `perf`, `test`, `ci`, `build`, `revert`

**Valid scopes:** `backend`, `frontend`, `infra`, `api`, `worker`, `ai`, `scanner`, `graph`, `ci`, `deps`

### Running Pre-commit Hooks

Hooks run automatically on commit:
```bash
git commit -m "feat(api): add new endpoint"
```

Run hooks manually on all files:
```bash
pre-commit run --all-files
```

Skip hooks (not recommended):
```bash
git commit --no-verify
```

### What's Being Checked

Pre-commit hooks mirror all CI checks:

**Go (backend/api, backend/worker):**
- `golangci-lint` - Linting and formatting
- `go mod tidy` - Dependency management
- `go build` - Build verification
- `gosec` - Security scanning

**Python (backend/ai):**
- `ruff` - Linting and formatting
- `mypy` - Type checking
- `bandit` - Security scanning

**Frontend:**
- `eslint` - Code linting
- `prettier` - Code formatting
- `npm audit` - Security audit (on CI only)

**General:**
- YAML validation
- Trailing whitespace removal
- End-of-file fixing
- Large file detection

## Support


For support, please open an issue in the standard GitHub issue tracker.

## License

This project is licensed under the MIT License - see the LICENSE file for details.