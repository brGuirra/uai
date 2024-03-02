# Go Version
FROM golang:1.22

# Install dependencies
RUN curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

RUN curl -L -o pkl https://github.com/apple/pkl/releases/download/0.25.2/pkl-alpine-linux-amd64 && \
  chmod +x pkl && \
  mv pkl /usr/bin/ && \
  /usr/bin/pkl --version

RUN curl -L https://github.com/go-task/task/releases/download/v3.31.0/task_linux_amd64.tar.gz | tar xvz && \
  mv task /usr/bin/ && \
  task --version

# Base setup for project
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY .. ./

# Build and start listening for file changes
ENTRYPOINT ["task", "dev"]
