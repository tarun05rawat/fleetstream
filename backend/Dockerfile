FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev pkgconfig

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o backend .

# Final lightweight runtime
FROM alpine:latest

# Install minimal runtime deps
RUN apk --no-cache add ca-certificates tzdata wget

# Use a consistent workdir
WORKDIR /app

# Copy binary
COPY --from=builder /app/backend .

# Expose required port
EXPOSE 8080

# Healthcheck (verify backend endpoint)
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

# Run as non-root (after workdir + binary copied)
RUN adduser -D -s /bin/sh factoryuser
USER factoryuser

CMD ["./backend"]
