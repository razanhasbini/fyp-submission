# Build stage
FROM golang:1.24-alpine AS build

# Install build dependencies for docconv (PDF parsing)
RUN apk add --no-cache \
    gcc \
    musl-dev \
    poppler-utils \
    tesseract-ocr \
    tesseract-ocr-data-eng

WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Update go.mod if needed
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# Production stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    poppler-utils \
    tesseract-ocr \
    tesseract-ocr-data-eng

WORKDIR /app

# Copy the binary from build stage
COPY --from=build /app/main .

# Expose the application port
EXPOSE 8089

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8089/v1/health || exit 1

# Run the application
CMD ["./main"]

