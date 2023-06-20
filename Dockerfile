# Start from a minimal base image with Go installed
FROM golang:alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download and cache the Go modules
RUN go mod download

# Copy the application source code
COPY . .

# Build the Go application
RUN go build -o user-service

# Start from a fresh Alpine Linux image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built executable from the builder stage
COPY --from=builder /app/user-service .
COPY --from=builder /app/.env . 

# Expose the port that the application listens on (if applicable)
EXPOSE 6000

# Run the application
CMD ["./user-service"]
