FROM python:3.11-slim
WORKDIR /app
COPY . .
RUN pip install uv
RUN uv sync
CMD ["sleep", "infinity"]