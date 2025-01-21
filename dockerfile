# Stage 1: Build the Go application
FROM golang:1.21-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./

# Download Go modules
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go application
RUN go build -o /app/main ./cmd/main.go

# Stage 2: Create a lightweight image to run the Go application
FROM alpine:latest

# Install any CA certificates needed for HTTPS
RUN apk --no-cache add ca-certificates

# Set the working directory inside the container
WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/main . 

# Ensure .env file is in the correct place
# COPY --from=builder /app/.env .

# Expose the application port (adjust if different)
EXPOSE 8082

# Run the application
CMD ["./main"]
