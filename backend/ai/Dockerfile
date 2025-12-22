FROM python:3.11-slim

WORKDIR /app
COPY . .

# Install uv
RUN pip install --no-cache-dir uv

# Use cache mount for uv to speed up dependency installation
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync

CMD ["uv", "run", "python", "-m", "app.main"]
