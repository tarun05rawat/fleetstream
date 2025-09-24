FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev pkgconfig

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sensor-simulator .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/sensor-simulator .

# Create a non-root user
RUN adduser -D -s /bin/sh sensor
USER sensor

# Expose health check endpoint (if implemented)
EXPOSE 8090

CMD ["./sensor-simulator"]