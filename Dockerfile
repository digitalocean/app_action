FROM golang:1.24.1-alpine

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o /usr/local/bin/deploy ./deploy && \
    go build -o /usr/local/bin/delete ./delete
