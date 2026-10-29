# Stage 1: Build static Go binary
FROM golang:alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/bin/graficalineal cmd/server/main.go

# Stage 2: Minimal runtime image
FROM alpine:3.20

WORKDIR /app

# Install TLS certificates for outbound HTTPS calls to DeepSeek API
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/bin/graficalineal /app/graficalineal

# Copy web static assets and templates
COPY --from=builder /app/web /app/web

# Expose default port
EXPOSE 8090

# Default environment variables
ENV PORT=8090
ENV DEEPSEEK_BASE_URL=https://api.deepseek.com
ENV DEEPSEEK_MODEL=deepseek-chat

# Run application
ENTRYPOINT ["/app/graficalineal"]
