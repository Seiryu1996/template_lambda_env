# Multi-stage build for Go Lambda
FROM golang:1.23-alpine AS builder

# Install necessary packages
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o bin/weather-lambda \
    ./cmd/weather-lambda

# Final stage
FROM public.ecr.aws/lambda/provided:al2-x86_64

# Copy the binary
COPY --from=builder /app/bin/weather-lambda ${LAMBDA_RUNTIME_DIR}/

# Set the entrypoint
CMD ["weather-lambda"]