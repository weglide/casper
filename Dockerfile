FROM golang:latest

WORKDIR /app
COPY ./ /app

# Install dependencies
RUN chmod +x scripts/dependencies.sh && ./scripts/dependencies.sh

# Set Environment variables
ENV LOCAL=True

# live reload
ENTRYPOINT CompileDaemon --build="go build main.go" --command=./main
