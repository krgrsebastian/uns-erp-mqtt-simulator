# Use the official Golang image as the build environment
FROM golang:1.18 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o mqtt_simulator .

# Start a new stage from scratch
FROM debian:buster

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/mqtt_simulator /app/mqtt_simulator

# Command to run the executable
ENTRYPOINT ["/app/mqtt_simulator"]
