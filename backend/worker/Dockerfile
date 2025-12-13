FROM golang:1.24-alpine
WORKDIR /app
COPY . .
RUN go mod init worker && go mod tidy || true
CMD ["sleep", "infinity"]
