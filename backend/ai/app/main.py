"""CloudCop AI Service - FastAPI + gRPC server."""

import threading
from contextlib import asynccontextmanager
from typing import AsyncIterator

from fastapi import FastAPI

from app.routers import health
from app.services.summarization import serve as grpc_serve


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    """Manage application lifecycle - start gRPC server."""
    # Start gRPC server in background thread
    grpc_server = grpc_serve(port=50051)
    grpc_server.start()

    def wait_for_termination() -> None:
        grpc_server.wait_for_termination()

    grpc_thread = threading.Thread(target=wait_for_termination, daemon=True)
    grpc_thread.start()

    yield

    # Cleanup
    grpc_server.stop(grace=5)


app = FastAPI(
    title="CloudCop AI Service",
    description="AI-powered security analysis and summarization",
    lifespan=lifespan,
)

app.include_router(health.router, prefix="/api")
