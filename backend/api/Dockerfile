FROM golang:1.22-alpine
WORKDIR /app
COPY . .
RUN go mod init api && go mod tidy || true
CMD ["sleep", "infinity"]