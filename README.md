# CloudCop Monorepo

This monorepo contains the CloudCop application, consisting of a Next.js frontend, backend services (Go GraphQL API, Go workers, FastAPI AI), infrastructure configurations, and deployment scripts.

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