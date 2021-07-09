# Start from the latest golang base image
FROM golang:1.16-alpine
# Set the Current Working Directory inside the container
WORKDIR /app
# Copy go mod and sum files
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download
RUN apk add -U curl git jq
# Install doctl
RUN export DOCTL_VERSION="$(curl https://github.com/digitalocean/doctl/releases/latest -s -L -I -o /dev/null -w '%{url_effective}' | awk '{n=split($1,A,"/v"); print A[n]}')" && \
    curl -sL https://github.com/digitalocean/doctl/releases/download/v$DOCTL_VERSION/doctl-$DOCTL_VERSION-linux-amd64.tar.gz | tar -xz -C /usr/local/bin && \
    chmod +x /usr/local/bin/doctl

# Copy the source from the current directory to the Working Directory inside the container
COPY . .
# Build the Go app
RUN go build -o main main.go
# Command to run the executable
RUN chmod +x main


ENTRYPOINT [ "/app/main" ]