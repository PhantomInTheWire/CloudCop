# Use python 3.13 to match pyproject.toml requirements (>=3.13)
FROM python:3.13-slim

WORKDIR /app

# Install uv
RUN pip install --no-cache-dir uv

# Copy dependency files first to allow layer caching
COPY pyproject.toml uv.lock ./

# Set uv link mode to copy to avoid hardlink warnings in Docker
ENV UV_LINK_MODE=copy

# Install dependencies
# --frozen ensures we use exact versions from uv.lock
# --no-install-project avoids installing the package itself (we just want deps)
# We mount:
# 1. /root/.cache/uv: for package cache
# 2. /root/.local/share/uv: for managed python versions (toolchains)
RUN --mount=type=cache,target=/root/.cache/uv \
    --mount=type=cache,target=/root/.local/share/uv \
    uv sync --frozen --no-install-project

# Copy the application source code
COPY . .

# Install the project and sync environment
RUN --mount=type=cache,target=/root/.cache/uv \
    --mount=type=cache,target=/root/.local/share/uv \
    uv sync --frozen

# Ensure the virtual environment is on PATH
ENV PATH="/app/.venv/bin:$PATH"
# Add current directory to PYTHONPATH to ensure absolute imports work
ENV PYTHONPATH="/app:$PYTHONPATH"

# Run the application
CMD ["python", "-m", "app.main"]
