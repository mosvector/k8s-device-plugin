# Use a minimal base image with Go for building
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Copy the source code into the container
COPY . .

# Build the plugin
RUN go mod tidy && go build -o k8s-device-plugin

# Use a minimal runtime image
FROM alpine:3.20

# Copy the compiled binary from the builder stage
COPY --from=builder /app/k8s-device-plugin /usr/local/bin/k8s-device-plugin

# Set the entrypoint to the plugin binary
ENTRYPOINT ["/usr/local/bin/k8s-device-plugin"]
