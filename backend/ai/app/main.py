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
    print("DEBUG: Starting gRPC server on port 50051...", flush=True)
    # Start gRPC server in background thread
    try:
        grpc_server = grpc_serve(port=50051)
        grpc_server.start()
        print("DEBUG: gRPC server started successfully on [::]:50051", flush=True)

        def wait_for_termination() -> None:
            print("DEBUG: Waiting for gRPC termination...", flush=True)
            grpc_server.wait_for_termination()
            print("DEBUG: gRPC termination wait ended", flush=True)

        grpc_thread = threading.Thread(target=wait_for_termination, daemon=True)
        grpc_thread.start()

        yield

        # Cleanup
        print("DEBUG: Stopping gRPC server...", flush=True)
        grpc_server.stop(grace=5)
        print("DEBUG: gRPC server stopped", flush=True)
    except Exception as e:
        print(f"DEBUG: Failed to start gRPC server: {e}", flush=True)
        logger.error(f"Failed to start gRPC server: {e}")
        raise


app = FastAPI(
    title="CloudCop AI Service",
    description="AI-powered security analysis and summarization",
    lifespan=lifespan,
)

app.include_router(health.router, prefix="/api")
