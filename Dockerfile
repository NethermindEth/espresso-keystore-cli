# Builder stage
FROM golang:1.22.2 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY --link . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/espresso-keystore-cli

# Final stage
FROM debian:buster-slim

# Install CA certificates
RUN apt-get update && apt-get install -y apt-transport-https ca-certificates gnupg curl

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/bin .

# Command to run the executable
CMD ["./espresso-keystore-cli"]
