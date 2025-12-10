# CloudCop API Service

Go GraphQL API service using Gin framework.

## Development

Install dependencies:
```bash
go mod tidy
```

Run locally:
```bash
go run ./cmd/server/main.go
```

## Health Check

GET `/health` - Returns service health status.