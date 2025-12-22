from fastapi import FastAPI

from app.routers import health

app = FastAPI(title="CloudCop AI Service")

app.include_router(health.router, prefix="/api")

# later: import dspy and init model registry here
