# CloudCop AI Service

FastAPI service with DSPy integration for AI-powered security analysis.

## Development

Install dependencies:
```bash
uv sync
```

Run locally:
```bash
uv run uvicorn app.main:app --reload
```

## Health Check

GET `/api/health` - Returns service health status.