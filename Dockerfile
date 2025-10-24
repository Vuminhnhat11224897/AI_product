# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pipeline main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/pipeline .

# Copy configuration files
COPY config/config.yaml ./config/
COPY prompts/ ./prompts/

# Create necessary directories
RUN mkdir -p data logs metadata

# Set timezone
ENV TZ=Asia/Ho_Chi_Minh

# Expose port (if needed for metrics/health checks in future)
EXPOSE 8080

# Run the pipeline
CMD ["./pipeline"]
