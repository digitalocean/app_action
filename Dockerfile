FROM golang:1.24.0-alpine AS build

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o /usr/local/bin/deploy ./deploy && \
    go build -o /usr/local/bin/delete ./delete

FROM cgr.dev/chainguard/static:latest

COPY --from=build /usr/local/bin/deploy /usr/local/bin/deploy
COPY --from=build /usr/local/bin/delete /usr/local/bin/delete
