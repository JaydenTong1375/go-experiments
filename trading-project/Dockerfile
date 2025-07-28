# Use official Golang image as builder
FROM golang:1.24-alpine AS builder

# Set working directory inside container
WORKDIR /app

# Copy the local Go file into the container
COPY . .

# Build the Go application
RUN go build -o server main.go

# Use a minimal image to run the binary
FROM alpine:latest

# Copy the built binary from the builder stage
COPY --from=builder /app/server /server
COPY --from=builder /app/Cert /Cert

# Expose the port your server listens on (adjust if needed)
EXPOSE 8090

# Command to run the binary
ENTRYPOINT ["/server"]
