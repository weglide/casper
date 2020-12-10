FROM golang:latest

WORKDIR /app

COPY ./ /app

# install dependencies (TODO: proper package manager)
RUN go get github.com/paulmach/orb
RUN go get github.com/gogo/protobuf/proto
RUN go get github.com/pkg/errors

RUN go get github.com/gorilla/mux

RUN go get github.com/lib/pq

RUN go get github.com/githubnemo/CompileDaemon

RUN go get github.com/aws/aws-lambda-go/lambda
RUN go get github.com/aws/aws-lambda-go/events

ENV LOCAL=True

# live reload
ENTRYPOINT CompileDaemon --build="go build main.go" --command=./main
