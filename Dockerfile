# Start from the latest golang base image
FROM golang:1.16-alpine as builder
# Set the Current Working Directory inside the container
WORKDIR /app
# Copy go mod and sum files
COPY . . 
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download
RUN apk add -U curl jq
# Install doctl
RUN export DOCTL_VERSION="$(curl https://github.com/digitalocean/doctl/releases/latest -s -L -I -o /dev/null -w '%{url_effective}' | awk '{n=split($1,A,"/v"); print A[n]}')" && \
    curl -sL https://github.com/digitalocean/doctl/releases/download/v$DOCTL_VERSION/doctl-$DOCTL_VERSION-linux-amd64.tar.gz | tar -xz -C /usr/local/bin && \
    chmod +x /usr/local/bin/doctl
# Build the Go app
RUN go build -mod=vendor -o bin/app_action
# # -- Stage 2 -- #
# # Create the final environment with the compiled binary.
# FROM alpine
# # Install any required dependencies.
# RUN apk --no-cache add ca-certificates
# WORKDIR /root/
# # Copy the binary from the builder stage and set it as the default command.
# COPY --from=builder /app/bin/app_action /usr/local/bin/

# Command to run the executable
ENTRYPOINT [ "app_action" ]