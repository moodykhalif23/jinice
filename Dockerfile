# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Copy web directory to /app/web
COPY web/ ./web/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/app

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary and web files from builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/web ./web/

# Expose port
EXPOSE 8080

# Command to run
CMD ["./main"]
