# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o chrome-service .

# Runtime stage
FROM chromedp/headless-shell:latest

# Install ghostscript for PDF/A conversion
USER root
RUN apt-get update && apt-get install -y \
    ghostscript \
    && rm -rf /var/lib/apt/lists/*

# Copy binary from builder
COPY --from=builder /app/chrome-service /usr/local/bin/chrome-service

# Create non-root user
RUN useradd -m -u 1000 appuser && \
    chown -R appuser:appuser /usr/local/bin/chrome-service

USER appuser

EXPOSE 8080

CMD ["/usr/local/bin/chrome-service"]
