# Use the official Golang image
FROM golang:1.21 as builder

# Set the working directory
WORKDIR /app

# Copy the source code
COPY . .

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /cache-server

# Use a minimal base image for the final stage
FROM alpine:latest

# Install CA certificates for HTTPS support
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /cache-server /cache-server

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["/cache-server"]
