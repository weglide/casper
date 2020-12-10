FROM golang:latest

WORKDIR /app
COPY ./ /app

# Install dependencies
RUN chmod +x dependencies.sh && ./dependencies.sh

# Set Environment variables
ENV LOCAL=True

# live reload
ENTRYPOINT CompileDaemon --build="go build main.go" --command=./main
