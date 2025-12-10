# CloudCop AI Service

FastAPI service with DSPy integration for AI-powered security analysis.

## Development

Install dependencies:
```bash
uv sync
```

Lint and format:
```bash
uv run ruff check app/
uv run ruff format app/
```

Type check:
```bash
uv run mypy app/
```

Run tests:
```bash
uv run pytest tests/
```

Run locally:
```bash
uv run uvicorn app.main:app --reload
```

## Health Check

GET `/api/health` - Returns service health status.