# CloudCop Worker Service

Go service for background job processing using Asynq.

## Development

Install dependencies:
```bash
go mod tidy
```

Run locally:
```bash
go run ./cmd/worker/main.go
```

## Health Check

The worker logs health status every 30 seconds.