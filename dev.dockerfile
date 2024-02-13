# Go Version
FROM golang:1.22-alpine

# Setup SHELL for Alpine
SHELL ["/bin/ash", "-o", "pipefail", "-c"]

# Environment variables which CompileDaemon requires to runs
ENV PROJECT_DIR=/app \
  GO111MODULE=on \
  CGO_ENABLED=0

# Base setup for project
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY .. ./

# Install dependencies
RUN apk --no-cache add curl && \
  go get github.com/githubnemo/CompileDaemon && \
  go install github.com/githubnemo/CompileDaemon && \
  curl -L https://github.com/go-task/task/releases/download/v3.31.0/task_linux_amd64.tar.gz | tar xvz && \
  mv task /usr/bin/

# Build and start listening for file changes
ENTRYPOINT ["task", "dev"]
