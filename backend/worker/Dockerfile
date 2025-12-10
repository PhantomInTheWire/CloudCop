FROM golang:1.23-alpine
WORKDIR /app
COPY . .
RUN go mod init worker && go mod tidy || true
CMD ["sleep", "infinity"]