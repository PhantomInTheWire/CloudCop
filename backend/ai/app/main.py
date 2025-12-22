"""CloudCop AI Service - FastAPI + gRPC server."""

import logging
import threading
from contextlib import asynccontextmanager
from typing import AsyncIterator

from fastapi import FastAPI

from app.routers import health
from app.services.summarization import serve as grpc_serve

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    """Manage application lifecycle - start gRPC server."""
    logger.info("Starting gRPC server on port 50051...")
    # Start gRPC server in background thread
    try:
        grpc_server = grpc_serve(port=50051)
        grpc_server.start()
        logger.info("gRPC server started successfully on [::]:50051")

        def wait_for_termination() -> None:
            grpc_server.wait_for_termination()

        grpc_thread = threading.Thread(target=wait_for_termination, daemon=True)
        grpc_thread.start()

        yield

        # Cleanup
        logger.info("Stopping gRPC server...")
        grpc_server.stop(grace=5)
        logger.info("gRPC server stopped")
    except Exception as e:
        logger.error(f"Failed to start gRPC server: {e}")
        raise


app = FastAPI(
    title="CloudCop AI Service",
    description="AI-powered security analysis and summarization",
    lifespan=lifespan,
)

app.include_router(health.router, prefix="/api")
