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

WORKDIR /app

# Copy the binary from builder stage and fix permissions
COPY --from=builder /app/sensor-simulator /app/sensor-simulator
RUN chmod +x /app/sensor-simulator && \
    adduser -D -s /bin/sh sensor && \
    chown sensor:sensor /app/sensor-simulator

USER sensor

EXPOSE 8090
CMD ["./sensor-simulator"]
